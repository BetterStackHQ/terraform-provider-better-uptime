package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceMetadata(t *testing.T) {
	server := createMetadataTestServer(t)
	defer server.Close()

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
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_metadata" "this" {
					owner_id   = "123"
					owner_type = "Monitor"
					key        = "test_key"
					metadata_value {
						value = "test_value"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_metadata.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_type", "Monitor"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "key", "test_key"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "metadata_value.0.value", "test_value"),
				),
			},
			// Step 2 - update.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_metadata" "this" {
					owner_id   = "123"
					owner_type = "Monitor"
					key        = "test_key"
					metadata_value {
						value = "test_value_changed"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_metadata.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_type", "Monitor"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "key", "test_key"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "metadata_value.0.value", "test_value_changed"),
				),
			},
			// Step 3 - update key, expect the original to be removed
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_metadata" "this" {
					owner_id   = "123"
					owner_type = "Monitor"
					key        = "test_key_changed"
					metadata_value {
						value = "test_value"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_metadata.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_type", "Monitor"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "key", "test_key_changed"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "metadata_value.0.value", "test_value"),
					server.TestCheckCalledRequest("POST", "/api/v3/metadata", `{"owner_type":"Monitor","owner_id":"123","key":"test_key","values":null}`),
				),
			},
			// Step 4 - make no changes, check plan is empty.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_metadata" "this" {
					owner_id   = "123"
					owner_type = "Monitor"
					key        = "test_key_changed"
					metadata_value {
						value = "test_value"
					}
				}`,
				PlanOnly: true,
			},
		},
	})
}

func createMetadataTestServer(t *testing.T) *TestServer {
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
			t.Fail()
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

		prefix := "/api/v3/metadata"
		id := "123456"

		switch {
		case r.Method == http.MethodPost && r.RequestURI == prefix:
			ts.Data.Store(body)
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
		case r.Method == http.MethodGet && r.RequestURI == fmt.Sprintf(`%s?owner_id=%s&owner_type=%s`, prefix, "123", "Monitor"):
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"999","attributes":{"key":"other","value":"other"}},{"id":%q,"attributes":%s}]}`, id, ts.Data.Load().([]byte))))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
			t.Fail()
		}
	}))

	return ts
}
