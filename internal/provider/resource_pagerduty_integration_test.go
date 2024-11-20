package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourcePagerDutyIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/pager-duty-webhooks", "1", "password")
	defer server.Close()

	var name = "test"
	var key = "keykeykeykey"
	var pdSeverity = "critical"

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

				resource "betteruptime_pagerduty_integration" "this" {
					name = "%s"
					key  = "%s"
					severity  = "%s"
				}
				`, name, key, pdSeverity),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_pagerduty_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_pagerduty_integration.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_pagerduty_integration.this", "key", key),
					resource.TestCheckResourceAttr("betteruptime_pagerduty_integration.this", "severity", pdSeverity),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_pagerduty_integration" "this" {
					name = "%s1"
					key  = "%s"
					severity  = "%s"
				}
				`, name, key, pdSeverity),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_pagerduty_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_pagerduty_integration.this", "name", fmt.Sprintf("%s1", name)),
					resource.TestCheckResourceAttr("betteruptime_pagerduty_integration.this", "key", key),
					resource.TestCheckResourceAttr("betteruptime_pagerduty_integration.this", "severity", pdSeverity),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_pagerduty_integration" "this" {
					name = "%s1"
					key  = "%s"
					severity  = "%s"
				}
				`, name, key, pdSeverity),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_pagerduty_integration.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
