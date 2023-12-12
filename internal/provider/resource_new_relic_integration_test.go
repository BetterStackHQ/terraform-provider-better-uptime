package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceNewRelicIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/new-relic-integrations", "1", "password")
	defer server.Close()

	var name = "test"
	var policy_id = "1234"

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

				resource "betteruptime_new_relic_integration" "this" {
					name	  = "%s"
					policy_id = %s
				}
				`, name, policy_id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_new_relic_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_new_relic_integration.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_new_relic_integration.this", "policy_id", policy_id),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_new_relic_integration" "this" {
					name	  = "%s1"
					policy_id = %s
				}
				`, name, policy_id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_new_relic_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_new_relic_integration.this", "name", fmt.Sprintf("%s1", name)),
					resource.TestCheckResourceAttr("betteruptime_new_relic_integration.this", "policy_id", policy_id),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_new_relic_integration" "this" {
					name	  = "%s1"
					policy_id = %s
				}
				`, name, policy_id),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_new_relic_integration.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
