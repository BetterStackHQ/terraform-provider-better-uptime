package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Received " + r.Method + " " + r.RequestURI)

		if r.Header.Get("Authorization") != "Bearer foo" {
			t.Fatal("Not authorized: " + r.Header.Get("Authorization"))
			t.Fail()
		}

		prefix := "/api/v2/policies"

		switch {
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=1":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"name": "Policy A", "incident_token":"abc","repeat_count":3, "repeat_delay":60}}],"pagination":{"next":"..."}}`))
		case r.Method == http.MethodGet && r.RequestURI == prefix+"?page=2":
			_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"name": "Policy B", "incident_token":"def","repeat_count":4, "repeat_delay":120}}],"pagination":{"next":null}}`))
		default:
			t.Fatal("Unexpected " + r.Method + " " + r.RequestURI)
			t.Fail()
		}
	}))
	defer server.Close()

	var name = "Policy B"

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

				data "betteruptime_policy" "this" {
					name = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betteruptime_policy.this", "id"),
					resource.TestCheckResourceAttr("data.betteruptime_policy.this", "name", name),
					resource.TestCheckResourceAttr("data.betteruptime_policy.this", "repeat_count", "4"),
					resource.TestCheckResourceAttr("data.betteruptime_policy.this", "repeat_delay", "120"),
					resource.TestCheckResourceAttr("data.betteruptime_policy.this", "incident_token", "def"),
				),
			},
		},
	})
}

// TODO: test duplicate
