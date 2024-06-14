package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type CalledRequest struct {
	Method string
	URL    string
	Body   string
}

type TestServer struct {
	*httptest.Server
	CalledRequests []CalledRequest
	mu   sync.Mutex
}

func newResourceServer(t *testing.T, baseRequestURI, id string, fieldsNotReturnedFromApi ...string) *TestServer {
	var data atomic.Value

	ts := &TestServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		ts.mu.Lock()
		ts.CalledRequests = append(ts.CalledRequests, CalledRequest{Method: r.Method, URL: r.RequestURI, Body: string(body)})
		ts.mu.Unlock()

		t.Log("Received " + r.Method + " " + r.RequestURI + " " + string(body))

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		switch {
		case r.Method == http.MethodPost && r.RequestURI == baseRequestURI:
			parsedBody := make(map[string]interface{})
			if err := json.Unmarshal(body, &parsedBody); err != nil {
				t.Fatal(err)
			}
			for _, field := range fieldsNotReturnedFromApi {
				delete(parsedBody, field)
			}
			body, err = json.Marshal(parsedBody)
			if err != nil {
				t.Fatal(err)
			}

			data.Store(body)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, body)))
		case r.Method == http.MethodGet && r.RequestURI == baseRequestURI+"/"+id:
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, data.Load().([]byte))))
		case r.Method == http.MethodPatch && r.RequestURI == baseRequestURI+"/"+id:
			patch := make(map[string]interface{})
			if err = json.Unmarshal(data.Load().([]byte), &patch); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(body, &patch); err != nil {
				t.Fatal(err)
			}
			patched, err := json.Marshal(patch)
			if err != nil {
				t.Fatal(err)
			}
			data.Store(patched)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, patched)))
		case r.Method == http.MethodDelete && r.RequestURI == baseRequestURI+"/"+id:
			w.WriteHeader(http.StatusNoContent)
			data.Store([]byte(nil))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))

	return ts
}

func (ts *TestServer) TestCheckCalledRequest(method, url, body string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		for _, req := range ts.CalledRequests {
			if req.Method == method && req.URL == url && req.Body == body {
				return nil
			}
		}
		return fmt.Errorf(`expected request %s %s with body "%s" not found`, method, url, body)
	}
}
