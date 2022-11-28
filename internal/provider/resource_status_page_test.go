package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceStatusPage(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages", "1", "password")
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
				    password     = "secret123"
				}
				`, subdomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "subdomain", subdomain),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "timezone", "UTC"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "password", "secret123"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "automatic_reports", "false"),
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
				    password     = "secret1234"
				}
				`, subdomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "subdomain", subdomain),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr("betteruptime_status_page.this", "password", "secret1234"),
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
				    password     = "secret1234"
				}
				`, subdomain),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_status_page.this",
				ImportState:       true,
				ImportStateVerify: true,
				// Password can't be imported and must be ignored when verifying imported state
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
