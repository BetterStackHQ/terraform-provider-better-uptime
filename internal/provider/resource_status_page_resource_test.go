package provider

import (
	"fmt"
	"regexp"
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
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_degraded_metadata_rule.#", "0"),
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
			// Step 5 - remove metadata
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
					mark_as_down_for = "any_incident"
					mark_as_degraded_for = "no_incident"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_for", "any_incident"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_down_metadata_rule.#", "0"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_degraded_for", "no_incident"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "mark_as_degraded_metadata_rule.#", "0"),
				),
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 6 - destroy.
			{
				ResourceName:      "betteruptime_status_page_resource.this",
				ImportState:       true,
				ImportStateId:     "0/1",
				ImportStateVerify: true,
				PreConfig: func() {
					t.Log("step 6")
				},
			},
		},
	})
}

func TestResourceStatusPageResourceManuallyTrackedItem(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create ManuallyTrackedItem without resource_id.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_type  = "ManuallyTrackedItem"
					public_name    = "Manual Item"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", "Manual Item"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_type", "ManuallyTrackedItem"),
					// Verify POST body does not contain resource_id.
					server.TestCheckCalledRequest("POST", "/api/v2/status-pages/0/resources", `{"resource_type":"ManuallyTrackedItem","public_name":"Manual Item","fixed_position":true}`),
				),
				PreConfig: func() {
					t.Log("step 1 - create ManuallyTrackedItem")
				},
			},
			// Step 2 - PlanOnly no-op to verify no drift.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_type  = "ManuallyTrackedItem"
					public_name    = "Manual Item"
				}
				`,
				PlanOnly: true,
				PreConfig: func() {
					t.Log("step 2 - PlanOnly no-op")
				},
			},
			// Step 3 - update public_name, verify PATCH body omits resource_id.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_type  = "ManuallyTrackedItem"
					public_name    = "Updated Manual Item"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "public_name", "Updated Manual Item"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/status-pages/0/resources/1", `{"public_name":"Updated Manual Item","fixed_position":true}`),
				),
				PreConfig: func() {
					t.Log("step 3 - update public_name")
				},
			},
			// Step 4 - import.
			{
				ResourceName:      "betteruptime_status_page_resource.this",
				ImportState:       true,
				ImportStateId:     "0/1",
				ImportStateVerify: true,
				PreConfig: func() {
					t.Log("step 4 - import")
				},
			},
		},
	})
}

func TestResourceStatusPageResourceValidation(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Monitor without resource_id should fail.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_type  = "Monitor"
					public_name    = "Bad Config"
				}
				`,
				ExpectError: regexp.MustCompile(`resource_id is required when resource_type is Monitor`),
			},
		},
	})
}
