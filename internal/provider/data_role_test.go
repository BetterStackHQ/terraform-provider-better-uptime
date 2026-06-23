package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestRoleDataSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v2/roles" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[
				{"id":"10","type":"role","attributes":{"name":"Team lead","role":"team_lead"}},
				{"id":"42","type":"role","attributes":{"name":"On-call lead","role":"custom"}}
			]}`))
			return
		}
		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

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
					provider "betteruptime" { api_token = "foo" }
					data "betteruptime_role" "oncall" { name = %q }
				`, "On-call lead"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.betteruptime_role.oncall", "id", "42"),
					resource.TestCheckResourceAttr("data.betteruptime_role.oncall", "role", "custom"),
				),
			},
		},
	})
}
