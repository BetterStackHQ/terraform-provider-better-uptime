package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceMonitor(t *testing.T) {
	var data atomic.Value
	server := createTestServer(t, &data)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					paused       = true
					regions      = ["us", "eu"]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "computed_by_betteruptime"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                = "%s"
					monitor_type       = "%s"
					pronounceable_name = "override"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "override"),
				),
			},
			// Step 3 - update (but preserve pronounceable_name).
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					http_method  = "POST"
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "http_method", "POST"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "override"),
				),
			},
			// Step 4 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					http_method  = "POST"
				}
				`, url, monitorType),
				PlanOnly: true,
			},
			// Step 5 - destroy.
			{
				ResourceName:      "betteruptime_monitor.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceMonitorWithHeaders(t *testing.T) {
	var data atomic.Value
	server := createTestServer(t, &data)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "status"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST"
							value = "test"
						}
					]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "request_headers.0.id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.name", "X-TEST"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.value", "test"),
				),
			},
			// Step 2 - add another request header.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST"
							value = "test"
						},
						{
							name  = "X-TEST-2"
							value = "test-2"
						}
					]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.name", "X-TEST"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.value", "test"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.1.name", "X-TEST-2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.1.value", "test-2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.#", "2"),
				),
			},
			// Step 2 - remove the first header.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url          = "%s"
					monitor_type = "%s"
					request_headers = [
						{
							name  = "X-TEST-2"
							value = "test-2"
						}
					]
				}
				`, url, monitorType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "request_headers.0.id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.name", "X-TEST-2"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "request_headers.0.value", "test-2"),
					resource.TestCheckNoResourceAttr("betteruptime_monitor.this", "request_headers.1.name"),
				),
			},
		},
	})
}

func TestExpectedStatusCodeMonitor(t *testing.T) {
	var data atomic.Value
	server := createTestServer(t, &data)
	defer server.Close()

	var url = "http://example.com"
	var monitorType = "expected_status_code"
	var expectedStatusCodes = []int{500, 501, 502}

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_monitor" "this" {
					url                   = "%s"
					monitor_type          = "%s"
					expected_status_codes = [%d, %d, %d]
					paused                = true
					regions               = ["us", "eu"]
				}
				`, url, monitorType, expectedStatusCodes[0], expectedStatusCodes[1], expectedStatusCodes[2]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_monitor.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "monitor_type", monitorType),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.#", strconv.Itoa(len(expectedStatusCodes))),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.0", strconv.Itoa(expectedStatusCodes[0])),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.1", strconv.Itoa(expectedStatusCodes[1])),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "expected_status_codes.2", strconv.Itoa(expectedStatusCodes[2])),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_monitor.this", "pronounceable_name", "computed_by_betteruptime"),
				),
			},
		},
	})
}

func createTestServer(t *testing.T, data *atomic.Value) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
			t.Fail()
		}

		prefix := "/api/v2/monitors"
		id := "1"

		switch {
		case r.Method == http.MethodPost && r.RequestURI == prefix:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			data.Store(body)
			w.WriteHeader(http.StatusCreated)
			// Inject pronounceable_name.
			computed := make(map[string]interface{})
			if err := json.Unmarshal(body, &computed); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			computed["pronounceable_name"] = "computed_by_betteruptime"
			if computed["request_headers"] != nil {
				computed["request_headers"] = processRequestHeaders(computed["request_headers"].([]interface{}))
			}
			body, err = json.Marshal(computed)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, body)))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"/"+id:
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, data.Load().([]byte))))
		case r.Method == http.MethodPatch && r.RequestURI == prefix+"/"+id:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			patch := make(map[string]interface{})
			if err = json.Unmarshal(data.Load().([]byte), &patch); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			if err = json.Unmarshal(body, &patch); err != nil {
				t.Fatal(err)
				t.Fail()
			}
			if patch["request_headers"] != nil {
				patch["request_headers"] = processRequestHeaders(patch["request_headers"].([]interface{}))
			}
			patched, err := json.Marshal(patch)
			if err != nil {
				t.Fatal(err)
				t.Fail()
			}
			data.Store(patched)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%q,"attributes":%s}}`, id, patched)))
		case r.Method == http.MethodDelete && r.RequestURI == prefix+"/"+id:
			w.WriteHeader(http.StatusNoContent)
			data.Store([]byte(nil))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
			t.Fail()
		}
	}))
}

func processRequestHeaders(headers []interface{}) []interface{} {
	var newHeaders []interface{}

	for _, v := range headers {
		header := v.(map[string]interface{})
		header["id"] = "1"

		if _, ok := header["_destroy"]; !ok {
			newHeaders = append(newHeaders, header)
		}
	}

	return newHeaders
}
