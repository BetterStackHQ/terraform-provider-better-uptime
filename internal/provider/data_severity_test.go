package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSeverity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
		}

		prefix := "/api/v2/urgencies"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id": "1", "type": "urgency", "attributes":{"name": "High Severity", "sms": false, "call": true, "email": true, "push": true}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id": "2", "type": "urgency", "attributes":{"name": "Low Severity", "sms": false, "call": false, "email": true, "push": true}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
		}
	}))
	defer server.Close()

	var name = "Low Severity"

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

				data "betteruptime_severity" "this" {
					name = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_severity.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_severity.this", "name", name),
					resource.TestCheckResourceAttr("data.betteruptime_severity.this", "sms", "false"),
					resource.TestCheckResourceAttr("data.betteruptime_severity.this", "call", "false"),
					resource.TestCheckResourceAttr("data.betteruptime_severity.this", "email", "true"),
					resource.TestCheckResourceAttr("data.betteruptime_severity.this", "push", "true"),
				),
			},
		},
	})
}

// TODO: test duplicate
