package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceStatusPageResource(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
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

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "2"
					resource_type  = "Monitor"
					public_name    = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", name),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "2"),
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

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "3"
					resource_type  = "Monitor"
					public_name    = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", name),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "3"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/status-pages/0/resources/1", `{"resource_id":3,"resource_type":"Monitor","fixed_position":true}`),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 3 - update with metadata rules.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "3"
					resource_type  = "Monitor"
					public_name    = "%s"
					mark_as_down_for = "incident_matching_metadata"
					mark_as_down_metadata_rule {
						key = "Default escalation policy"
						metadata_value {
							type = "Policy"
							item_id = "102683"
						}
						metadata_value {
							type = "Policy"
							item_id = "89964"
						}
					}
					mark_as_degraded_for = "any_incident"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", name),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "3"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_for", "incident_matching_metadata"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_metadata_rule.0.key", "Default escalation policy"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_metadata_rule.0.metadata_value.0.type", "Policy"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_metadata_rule.0.metadata_value.0.item_id", "102683"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_metadata_rule.0.metadata_value.1.type", "Policy"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_metadata_rule.0.metadata_value.1.item_id", "89964"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_degraded_for", "any_incident"),
				),
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 4 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "3"
					resource_type  = "Monitor"
					public_name    = "%s"
					mark_as_down_for = "incident_matching_metadata"
					mark_as_down_metadata_rule {
						key = "Default escalation policy"
						metadata_value {
							type = "Policy"
							item_id = "102683"
						}
						metadata_value {
							type = "Policy"
							item_id = "89964"
						}
					}
					mark_as_degraded_for = "any_incident"
				}
				`, name),
				PlanOnly: true,
				PreConfig: func() {
					t.Log("step 4")
				},
			},
			// Step 5 - destroy.
			{
				ResourceName:      "betteruptime_status_page_resource.this",
				ImportState:       true,
				ImportStateId:     "0/1",
				ImportStateVerify: false,
				PreConfig: func() {
					t.Log("step 5")
				},
			},
		},
	})
}
