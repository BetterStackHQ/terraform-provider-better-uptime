package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestResourceStatusPageResourceUnknownStatusHistoryField(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
	defer server.Close()

	// The API returns status_history entries with fields unknown to the provider
	// (e.g. degraded_duration, which caused https://github.com/BetterStackHQ/terraform-provider-better-uptime/issues/222).
	attributes := `{"resource_id":2,"resource_type":"Monitor","public_name":"example","status_history":[{"day":"2026-07-08","status":"downtime","downtime_duration":60,"maintenance_duration":0,"degraded_duration":30,"another_future_field":"ignored"}]}`
	server.ExpectRequest("POST", "/api/v2/status-pages/0/resources", "", 201, `{"data":{"id":"1","attributes":`+attributes+`}}`)
	server.ExpectRequest("GET", "/api/v2/status-pages/0/resources/1", "", 200, `{"data":{"id":"1","attributes":`+attributes+`}}`)

	config := `
	provider "betteruptime" {
		api_token = "foo"
	}

	resource "betteruptime_status_page_resource" "this" {
		status_page_id = "0"
		resource_id    = "2"
		resource_type  = "Monitor"
		public_name    = "example"
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create; unknown status_history fields must be ignored, known ones kept.
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_status_page_resource.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "status_history.#", "1"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "status_history.0.day", "2026-07-08"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "status_history.0.status", "downtime"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "status_history.0.downtime_duration", "60"),
					resource.TestCheckNoResourceAttr("betteruptime_status_page_resource.this", "status_history.0.degraded_duration"),
					resource.TestCheckNoResourceAttr("betteruptime_status_page_resource.this", "status_history.0.another_future_field"),
				),
			},
			// Step 2 - refresh + plan must not fail on the unknown fields either.
			{
				Config:   config,
				PlanOnly: true,
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

func TestResourceStatusPageResourceCatalogReference(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
	defer server.Close()

	// The API finds or creates the CatalogReference and returns its id as resource_id.
	payment := `{"data":{"id":"1","attributes":{"resource_id":77,"resource_type":"CatalogReference","public_name":"Payments","widget_type":"history","catalog_reference":{"key":"service","value":{"type":"String","value":"payment"}}}}}`
	platform := `{"data":{"id":"1","attributes":{"resource_id":78,"resource_type":"CatalogReference","public_name":"Payments","widget_type":"history","catalog_reference":{"key":"service","value":{"type":"Team","item_id":5,"name":"Platform"}}}}}`
	server.ExpectRequest("POST", "/api/v2/status-pages/0/resources", "", 201, payment)
	server.ExpectRequest("GET", "/api/v2/status-pages/0/resources/1", "", 200, payment)
	server.ExpectRequest("PATCH", "/api/v2/status-pages/0/resources/1", "", 200, platform)

	stringConfig := `
	provider "betteruptime" {
		api_token = "foo"
	}

	resource "betteruptime_status_page_resource" "this" {
		status_page_id = "0"
		resource_type  = "CatalogReference"
		public_name    = "Payments"
		catalog_reference {
			key = "service"
			metadata_value {
				value = "payment"
			}
		}
	}
	`
	teamConfig := `
	provider "betteruptime" {
		api_token = "foo"
	}

	resource "betteruptime_status_page_resource" "this" {
		status_page_id = "0"
		resource_type  = "CatalogReference"
		public_name    = "Payments"
		catalog_reference {
			key = "service"
			metadata_value {
				type = "Team"
				name = "Platform"
			}
		}
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create from a String catalog value, without resource_id.
			{
				Config: stringConfig,
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("POST", "/api/v2/status-pages/0/resources", `{"resource_type":"CatalogReference","public_name":"Payments","fixed_position":true,"catalog_reference":{"key":"service","value":{"type":"String","value":"payment"}}}`),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "77"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.key", "service"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.metadata_value.0.type", "String"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.metadata_value.0.value", "payment"),
				),
			},
			// Step 2 - no drift.
			{
				Config:   stringConfig,
				PlanOnly: true,
			},
			// Step 3 - repoint at a Team by name; the update sends no resource_id.
			{
				Config: teamConfig,
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("PATCH", "/api/v2/status-pages/0/resources/1", `{"resource_type":"CatalogReference","fixed_position":true,"catalog_reference":{"key":"service","value":{"type":"Team","name":"Platform"}}}`),
					server.TestCheckCalledRequestWithout("PATCH", "/api/v2/status-pages/0/resources/1", "resource_id"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "resource_id", "78"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.metadata_value.0.type", "Team"),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.metadata_value.0.name", "Platform"),
					// The API also returns item_id, but the user looked the team up by name.
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.metadata_value.0.item_id", ""),
					server.ReplaceExpectedResponseAfterApply("GET", "/api/v2/status-pages/0/resources/1", platform),
				),
			},
			// Step 4 - no drift from the item_id the API returns.
			{
				Config:   teamConfig,
				PlanOnly: true,
			},
			// Step 5 - import keeps the item_id, as there is no configured lookup field to follow.
			{
				ResourceName:            "betteruptime_status_page_resource.this",
				ImportState:             true,
				ImportStateId:           "0/1",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"catalog_reference.0.metadata_value.0.name", "catalog_reference.0.metadata_value.0.item_id"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if got := states[0].Attributes["catalog_reference.0.metadata_value.0.item_id"]; got != "5" {
						return fmt.Errorf("expected imported item_id 5, got %q", got)
					}
					return nil
				},
			},
		},
	})
}

func TestResourceStatusPageResourceCatalogReferenceByResourceID(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
	defer server.Close()

	// A CatalogReference row declared by resource_id still gets catalog_reference back from the API.
	response := `{"data":{"id":"1","attributes":{"resource_id":77,"resource_type":"CatalogReference","public_name":"Payments","widget_type":"history","catalog_reference":{"key":"service","value":{"type":"String","value":"payment"}}}}}`
	server.ExpectRequest("POST", "/api/v2/status-pages/0/resources", "", 201, response)
	server.ExpectRequest("GET", "/api/v2/status-pages/0/resources/1", "", 200, response)

	config := `
	provider "betteruptime" {
		api_token = "foo"
	}

	resource "betteruptime_status_page_resource" "this" {
		status_page_id = "0"
		resource_id    = "77"
		resource_type  = "CatalogReference"
		public_name    = "Payments"
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					server.TestCheckCalledRequest("POST", "/api/v2/status-pages/0/resources", `{"resource_id":77,"resource_type":"CatalogReference","public_name":"Payments","fixed_position":true}`),
					resource.TestCheckResourceAttr("betteruptime_status_page_resource.this", "catalog_reference.0.metadata_value.0.value", "payment"),
				),
			},
			// The computed catalog_reference must not show as a diff against a config without it.
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestResourceStatusPageResourceCatalogReferenceValidation(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-pages/0/resources", "1")
	defer server.Close()

	config := func(attributes string) string {
		return `
		provider "betteruptime" {
			api_token = "foo"
		}

		resource "betteruptime_status_page_resource" "this" {
			status_page_id = "0"
			public_name    = "Bad Config"
			` + attributes + `
		}
		`
	}
	stringReference := `
			catalog_reference {
				key = "service"
				metadata_value {
					value = "payment"
				}
			}`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config:      config(`resource_type = "CatalogReference"`),
				ExpectError: regexp.MustCompile(`resource_id or catalog_reference is required when resource_type is CatalogReference`),
			},
			{
				Config:      config(`resource_type = "Monitor"` + "\n" + `resource_id = "2"` + stringReference),
				ExpectError: regexp.MustCompile(`conflicts with`),
			},
			{
				Config:      config(`resource_type = "Monitor"` + stringReference),
				ExpectError: regexp.MustCompile(`catalog_reference can only be used when resource_type is CatalogReference, not Monitor`),
			},
			{
				Config: config(`resource_type = "CatalogReference"
			catalog_reference {
				key = "owner"
				metadata_value {
					type = "Team"
				}
			}`),
				ExpectError: regexp.MustCompile(`at least one of item_id, email, or name must be set for Team type`),
			},
		},
	})
}

func TestResourceStatusPageResourceSwitchToCatalogReference(t *testing.T) {
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
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_id    = "2"
					resource_type  = "Monitor"
					public_name    = "Payments"
				}
				`,
			},
			// The monitor's resource_id left in state must not be sent along with catalog_reference.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_status_page_resource" "this" {
					status_page_id = "0"
					resource_type  = "CatalogReference"
					public_name    = "Payments"
					catalog_reference {
						key = "service"
						metadata_value {
							value = "payment"
						}
					}
				}
				`,
				Check: server.TestCheckCalledRequest("PATCH", "/api/v2/status-pages/0/resources/1", `{"resource_type":"CatalogReference","fixed_position":true,"catalog_reference":{"key":"service","value":{"type":"String","value":"payment"}}}`),
			},
		},
	})
}
