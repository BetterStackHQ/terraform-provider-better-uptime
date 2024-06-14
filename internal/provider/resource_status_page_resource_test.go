package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceStatusPageResource(t *testing.T) {
	current_resource_id := -1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request_body := ""
		if body, err := io.ReadAll(r.Body); err == nil {
			request_body = string(body)
		}
		t.Log("Received " + r.Method + " " + r.RequestURI + " " + request_body)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
			t.Fail()
		}

		switch {
		case r.Method == http.MethodPost && r.RequestURI == "/api/v2/status-pages/0/resources" && request_body == `{"resource_id":2,"resource_type":"Monitor","public_name":"example","fixed_position":true}`:
			w.WriteHeader(http.StatusCreated)
			current_resource_id = 2
			_, _ = w.Write([]byte(`{"data":{"id":"123","type":"resource","attributes":{"resource_id":2,"resource_type":"Monitor","public_name":"example","explanation":"","history":true,"widget_type":"history","position":0,"availability":1.0,"status_history": []}}}`))
		case r.Method == http.MethodPatch && r.RequestURI == "/api/v2/status-pages/0/resources/123" && request_body == `{"resource_id":3,"resource_type":"Monitor","fixed_position":true}`:
			w.WriteHeader(http.StatusOK)
			current_resource_id = 3
			_, _ = w.Write([]byte(`{"data":{"id":"123","type":"resource","attributes":{"resource_id":3,"resource_type":"Monitor","public_name":"example","explanation":"","history":true,"widget_type":"history","position":0,"availability":1.0,"status_history": []}}}`))
		case r.Method == http.MethodDelete && r.RequestURI == "/api/v2/status-pages/0/resources/123" && request_body == ``:
			w.WriteHeader(http.StatusNoContent)
			current_resource_id = -1
		case r.Method == http.MethodGet && r.RequestURI == "/api/v2/status-pages/0/resources/123" && request_body == ``:
			_, _ = w.Write([]byte(`{"data":{"id":"123","type":"resource","attributes":{"resource_id":` + fmt.Sprint(current_resource_id) + `,"resource_type":"Monitor","public_name":"example","explanation":"","history":true,"widget_type":"history","position":0,"availability":1.0,"status_history": []}}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI + " " + request_body)
			t.Fail()
		}
	}))
	defer server.Close()

	var name = "example"

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

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "2"
					resource_type  = "Monitor"
					public_name    = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", name),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "2"),
				),
				PreConfig: func() {
					t.Log("step 1")
				},
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "3"
					resource_type  = "Monitor"
					public_name    = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", name),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "3"),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "3"
					resource_type  = "Monitor"
					public_name    = "%s"
				}
				`, name),
				PlanOnly: true,
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 4 - destroy.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"x
				}`,
				PreConfig: func() {
					t.Log("step 4")
				},
			},
		},
	})
}
