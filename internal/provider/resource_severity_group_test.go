package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceSeverityGroup(t *testing.T) {
	server := newResourceServer(t, "/api/v2/urgency-groups", "1")
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

				resource "betteruptime_severity_group" "this" {
					name       = "%s"
					sort_index = 1
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_severity_group.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_severity_group.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_severity_group.this", "sort_index", "1"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity_group" "this" {
					name       = "%s"
					sort_index = 0
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_severity_group.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_severity_group.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_severity_group.this", "sort_index", "0"),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity_group" "this" {
					name       = "%s"
					sort_index = 0
				}
				`, name),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_severity_group.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
