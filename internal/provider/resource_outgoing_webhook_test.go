package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceOutgoingWebhookIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/outgoing-webhooks", "1")
	defer server.Close()

	var name = "test"
	var url = "https://example.com/webhook"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_outgoing_webhook" "this" {
					name = "%s"
					url  = "%s"
					trigger_type = "incident_change"
					on_incident_started = true
					on_incident_acknowledged = true
					on_incident_resolved = true
				}
				`, name, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_outgoing_webhook.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "url", url),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "trigger_type", "incident_change"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "on_incident_started", "true"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "on_incident_acknowledged", "true"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "on_incident_resolved", "true"),
				),
			},
			// Step 2 - update
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_outgoing_webhook" "this" {
					name = "%s1"
					url  = "%s?different=true"
					trigger_type = "incident_change"
					on_incident_started = false
					on_incident_acknowledged = true
					on_incident_resolved = true
				}
				`, name, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_outgoing_webhook.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "name", fmt.Sprintf("%s1", name)),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "url", fmt.Sprintf("%s?different=true", url)),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.this", "on_incident_started", "false"),
				),
			},
			// Step 3 - make no changes, check plan is empty
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_outgoing_webhook" "this" {
					name = "%s1"
					url  = "%s?different=true"
					trigger_type = "incident_change"
					on_incident_started = false
					on_incident_acknowledged = true
					on_incident_resolved = true
				}
				`, name, url),
				PlanOnly: true,
			},
			// Step 4 - import
			{
				ResourceName:      "betteruptime_outgoing_webhook.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceOutgoingWebhookIntegrationCustomTemplate(t *testing.T) {
	server := newResourceServer(t, "/api/v2/outgoing-webhooks", "1", "password")
	defer server.Close()

	var name = "test-custom"
	var url = "https://example.com/webhook"
	var bodyTemplate = `{"incident":{"id":"$INCIDENT_ID","started_at":"$STARTED_AT"}}`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create with custom template
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_outgoing_webhook" "custom" {
					name = "%s"
					url  = "%s"
					trigger_type = "incident_change"

					custom_webhook_template_attributes {
						http_method = "patch"
						auth_user = "user"
						auth_password = "password"

						headers_template {
							name = "Content-Type"
							value = "application/json"
						}
						headers_template {
							name = "X-Custom-Header"
							value = "custom-value"
						}

						body_template = jsonencode(%s)
					}
				}
				`, name, url, bodyTemplate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_outgoing_webhook.custom", "id"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "name", name),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "url", url),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.http_method", "patch"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.auth_user", "user"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.auth_password", "password"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.headers_template.0.name", "Content-Type"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.headers_template.0.value", "application/json"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.headers_template.1.name", "X-Custom-Header"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.headers_template.1.value", "custom-value"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.body_template", bodyTemplate),
				),
			},
			// Step 2 - update template
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_outgoing_webhook" "custom" {
					name = "%s"
					url  = "%s"
					trigger_type = "incident_change"

					custom_webhook_template_attributes {
						http_method = "put"
						auth_user = "user2"
						auth_password = "password2"

						headers_template {
							name = "Content-Type"
							value = "application/json"
						}

						body_template = jsonencode(%s)
					}
				}
				`, name, url, bodyTemplate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_outgoing_webhook.custom", "id"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.http_method", "put"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.auth_user", "user2"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.auth_password", "password2"),
					resource.TestCheckResourceAttr("betteruptime_outgoing_webhook.custom", "custom_webhook_template_attributes.0.headers_template.#", "1"),
				),
			},
			// Step 3 - make no changes, check plan is empty
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_outgoing_webhook" "custom" {
					name = "%s"
					url  = "%s"
					trigger_type = "incident_change"

					custom_webhook_template_attributes {
						http_method = "put"
						auth_user = "user2"
						auth_password = "password2"

						headers_template {
							name = "Content-Type"
							value = "application/json"
						}

						body_template = jsonencode(%s)
					}
				}
				`, name, url, bodyTemplate),
				PlanOnly: true,
			},
			// Step 4 - import
			{
				ResourceName:      "betteruptime_outgoing_webhook.custom",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
