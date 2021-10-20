package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceStatusPageSection(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/sections", "1")
	defer server.Close()

	var name = "example"
	var updatedName = "example2"

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

				resource "betteruptime_status_page_section" "this" {
					status_page_id = "0"
					name           = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_section.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_section.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_status_page_section.this", "position", "0"),
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

				resource "betteruptime_status_page_section" "this" {
					status_page_id = "0"
					name           = "%s"
				}
				`, updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_section.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_section.this", "name", updatedName),
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

				resource "betteruptime_status_page_section" "this" {
					status_page_id = "0"
					name           = "%s"
					position       = "0"
				}
				`, updatedName),
				PlanOnly: true,
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 4 - destroy.
			{
				ResourceName:            "betteruptime_status_page_section.this",
				ImportState:             true,
				ImportStateId:           "0/1",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fixed_position"},
				PreConfig: func() {
					t.Log("step 4")
				},
			},
		},
	})
}
