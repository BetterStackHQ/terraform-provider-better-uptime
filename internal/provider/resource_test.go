package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// A provider block pointing at a test server, for tests that assemble their config by
// concatenation rather than inlining the whole thing.
const testProviderBlock = `
provider "betteruptime" {
  api_token = "foo"
}
`

func testProviderFactories(url string) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"betteruptime": func() (*schema.Provider, error) {
			return New(WithURL(url)), nil
		},
	}
}

type CalledRequest struct {
	Method string
	URL    string
	Body   string
}

type ExpectedRequest struct {
	Method     string
	URL        string
	Body       string
	StatusCode int
	Response   string
}

type TestServer struct {
	*httptest.Server
	CalledRequests   []CalledRequest
	ExpectedRequests []ExpectedRequest
	mu               sync.Mutex
	Data             atomic.Value
}

func newResourceServer(t *testing.T, baseRequestURI, id string, fieldsNotReturnedFromApi ...string) *TestServer {
	ts := &TestServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
			t.Fail()
		}
		ts.mu.Lock()
		ts.CalledRequests = append(ts.CalledRequests, CalledRequest{Method: r.Method, URL: r.RequestURI, Body: string(body)})
		ts.mu.Unlock()

		t.Log("Received " + r.Method + " " + r.RequestURI + " " + string(body))

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		// Check for custom request handlers first
		ts.mu.Lock()
		for _, expected := range ts.ExpectedRequests {
			if expected.Method == r.Method && expected.URL == r.RequestURI {
				if expected.Body == "" || expected.Body == string(body) {
					w.WriteHeader(expected.StatusCode)
					_, _ = w.Write([]byte(expected.Response))
					ts.mu.Unlock()
					return
				}
			}
		}
		ts.mu.Unlock()

		// Fall back to default handlers
		switch {
		case r.Method == http.MethodPost && r.RequestURI == baseRequestURI:
			parsedBody := make(map[string]interface{})
			if err := json.Unmarshal(body, &parsedBody); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			for _, field := range fieldsNotReturnedFromApi {
				delete(parsedBody, field)
			}
			body, err = json.Marshal(parsedBody)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}

			ts.Data.Store(body)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, body)))
		case r.Method == http.MethodGet && r.RequestURI == baseRequestURI+"/"+id:
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, ts.Data.Load().([]byte))))
		case r.Method == http.MethodPatch && r.RequestURI == baseRequestURI+"/"+id:
			patch := make(map[string]interface{})
			if err = json.Unmarshal(ts.Data.Load().([]byte), &patch); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			if err = json.Unmarshal(body, &patch); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			patched, err := json.Marshal(patch)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			ts.Data.Store(patched)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, patched)))
		case r.Method == http.MethodDelete && r.RequestURI == baseRequestURI+"/"+id:
			w.WriteHeader(http.StatusNoContent)
			ts.Data.Store([]byte(nil))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))

	return ts
}

func (ts *TestServer) ExpectRequest(method, url, body string, statusCode int, response string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.ExpectedRequests = append(ts.ExpectedRequests, ExpectedRequest{
		Method:     method,
		URL:        url,
		Body:       body,
		StatusCode: statusCode,
		Response:   response,
	})
}

// Asserts that a request was made and that its body does NOT mention needle. Useful for a key
// whose mere presence changes the API's behaviour, such as a null field the API reads as "destroy".
func (ts *TestServer) TestCheckCalledRequestWithout(method, url, needle string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		found := false
		for _, req := range ts.CalledRequests {
			if req.Method == method && req.URL == url {
				found = true
				if strings.Contains(req.Body, needle) {
					return fmt.Errorf(`request %s %s should not mention %q, got body "%s"`, method, url, needle, req.Body)
				}
			}
		}
		if !found {
			return fmt.Errorf("expected request %s %s not found", method, url)
		}
		return nil
	}
}

func (ts *TestServer) TestCheckCalledRequest(method, url, body string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		for _, req := range ts.CalledRequests {
			if req.Method == method && req.URL == url && (body == "" || req.Body == body) {
				return nil
			}
		}
		return fmt.Errorf(`expected request %s %s with body "%s" not found`, method, url, body)
	}
}
