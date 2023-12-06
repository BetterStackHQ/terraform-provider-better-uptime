package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceSplunkOnCallIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/splunk-on-calls", "1", "password")
	defer server.Close()

	var name = "test"
	var url = "https://alert.victorops.com/integrations/generic/0/alert/0/your_routing_key"

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

				resource "betteruptime_splunk_oncall_integration" "this" {
					name = "%s"
					url  = "%s"
				}
				`, name, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_splunk_oncall_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_splunk_oncall_integration.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_splunk_oncall_integration.this", "url", url),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_splunk_oncall_integration" "this" {
					name = "%s1"
					url  = "%s"
				}
				`, name, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_splunk_oncall_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_splunk_oncall_integration.this", "name", fmt.Sprintf("%s1", name)),
					resource.TestCheckResourceAttr("betteruptime_splunk_oncall_integration.this", "url", url),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_splunk_oncall_integration" "this" {
					name = "%s1"
					url  = "%s"
				}
				`, name, url),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_splunk_oncall_integration.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
