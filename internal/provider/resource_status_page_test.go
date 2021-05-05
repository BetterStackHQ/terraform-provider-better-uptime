package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceStatusPage(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages", "1")
	defer server.Close()

	var subdomain = "example"

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

				resource "betteruptime_status_page" "this" {
				    company_name = "Example, Inc"
				    company_url  = "https://example.com"
				    timezone     = "UTC"
				    subdomain    = "%s"
				}
				`, subdomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "subdomain", subdomain),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "timezone", "UTC"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page" "this" {
				    company_name = "Example, Inc"
				    company_url  = "https://example.com"
				    timezone     = "America/Los_Angeles"
				    subdomain    = "%s"
				}
				`, subdomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "subdomain", subdomain),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "timezone", "America/Los_Angeles"),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page" "this" {
				    company_name = "Example, Inc"
				    company_url  = "https://example.com"
				    timezone     = "America/Los_Angeles"
				    subdomain    = "%s"
				}
				`, subdomain),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_status_page.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
