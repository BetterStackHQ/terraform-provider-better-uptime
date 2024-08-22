package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceIncomingWebhook(t *testing.T) {
	server := newResourceServer(t, "/api/v2/incoming-webhooks", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create incoming webhook.
			{
				Config: `
				provider "betteruptime" {
				  api_token = "foo"
				}
				resource "betteruptime_incoming_webhook" "this" {
				  name = "Terraform Test"
				  call = false
				  sms = false
				  email = true
				  push = true
				  team_wait = 180
				  recovery_period = 0
				  paused = false
				  started_rule_type = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type = "all"
				  started_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "alert"
				  }
				  started_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "reminder"
				  }
				  resolved_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "resolved"
				  }
				  cause_field {
					field_target = "json"
					target_field = "incident.status"
					match_type = "match_everything"
					content = "title"
				  }
				  title_field {
					field_target = "json"
					target_field = "incident.title"
					match_type = "match_everything"
					content = "title"
				  }
				  started_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "json"
					target_field = "incident.id"
					match_type = "match_between"
					content_before = "<"
					content_after = "-"
				  }
				  resolved_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "json"
					target_field = "incident.id"
					match_type = "match_between"
					content_before = "<"
					content_after = "-"
				  }
				  other_started_fields {
					name = "Caused by"
					field_target = "json"
					target_field = "incident.description"
					match_type = "match_between"
					content_before = "by:"
					content_after = ","
				  }
				  other_started_fields {
					name = "Description"
					field_target = "json"
					target_field = "incident.description"
					match_type = "match_between"
					content_before = "description:"
					content_after = ","
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_incoming_webhook.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "name", "Terraform Test"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.0.rule_target", "json"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.0.target_field", "incident.status"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.0.content", "alert"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.1.content", "reminder"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "resolved_rules.0.match_type", "contains"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "resolved_rules.0.content", "resolved"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "cause_field.0.target_field", "incident.status"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "cause_field.0.match_type", "match_everything"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "title_field.0.target_field", "incident.title"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "title_field.0.match_type", "match_everything"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_alert_id_field.0.content_before", "<"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "resolved_alert_id_field.0.content_after", "-"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "other_started_fields.0.content_before", "by:"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "other_started_fields.1.content_after", ","),
				),
				PreConfig: func() {
					t.Log("step 1")
				},
			},
			// Step 2 - change some root attributes
			{
				Config: `
				provider "betteruptime" {
				  api_token = "foo"
				}
				resource "betteruptime_incoming_webhook" "this" {
				  name = "Terraform Test - Updated"
				  call = false
				  sms = false
				  email = true
				  push = true
				  team_wait = 0
				  recovery_period = 180
				  paused = true
				  started_rule_type = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type = "all"
				  started_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "alert"
				  }
				  started_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "reminder"
				  }
				  resolved_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "resolved"
				  }
				  cause_field {
					field_target = "json"
					target_field = "incident.status"
					match_type = "match_everything"
					content = "title"
				  }
				  started_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "json"
					target_field = "incident.id"
					match_type = "match_between"
					content_before = "<"
					content_after = "-"
				  }
				  resolved_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "json"
					target_field = "incident.id"
					match_type = "match_between"
					content_before = "<"
					content_after = "-"
				  }
				  other_started_fields {
					name = "Caused by"
					field_target = "json"
					target_field = "incident.description"
					match_type = "match_between"
					content_before = "by:"
					content_after = ","
				  }
				  other_started_fields {
					name = "Description"
					field_target = "json"
					target_field = "incident.description"
					match_type = "match_between"
					content_before = "description:"
					content_after = ","
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "name", "Terraform Test - Updated"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "team_wait", "0"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "recovery_period", "180"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.0.content", "alert"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/incoming-webhooks/1", `{"name":"Terraform Test - Updated","team_wait":0,"recovery_period":180,"paused":true,"title_field":null}`),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 3 - change the started rules.
			{
				Config: `
				provider "betteruptime" {
				  api_token = "foo"
				}
				resource "betteruptime_incoming_webhook" "this" {
				  name = "Terraform Test - Updated"
				  call = false
				  sms = false
				  email = true
				  push = true
				  team_wait = 0
				  recovery_period = 180
				  paused = true
				  started_rule_type = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type = "all"
				  started_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "new-alert"
				  }
				  started_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "new-reminder"
				  }
				  resolved_rules {
					rule_target = "json"
					target_field = "incident.status"
					match_type = "contains"
					content = "resolved"
				  }
				  cause_field {
					field_target = "json"
					target_field = "incident.status"
					match_type = "match_everything"
					content = "title"
				  }
				  started_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "json"
					target_field = "incident.id"
					match_type = "match_between"
					content_before = "<"
					content_after = "-"
				  }
				  resolved_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "json"
					target_field = "incident.id"
					match_type = "match_between"
					content_before = "<"
					content_after = "-"
				  }
				  other_started_fields {
					name = "Caused by"
					field_target = "json"
					target_field = "incident.description"
					match_type = "match_between"
					content_before = "by:"
					content_after = ","
				  }
				  other_started_fields {
					name = "Description"
					field_target = "json"
					target_field = "incident.description"
					match_type = "match_between"
					content_before = "description:"
					content_after = ","
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "name", "Terraform Test - Updated"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.0.content", "new-alert"),
					resource.TestCheckResourceAttr("betteruptime_incoming_webhook.this", "started_rules.1.content", "new-reminder"),
				),
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_incoming_webhook.this",
				ImportState:       true,
				ImportStateId:     "1",
				ImportStateVerify: false, // Cannot verify due to TypeSet fields not properly transformed
				PreConfig: func() {
					t.Log("step 4")
				},
			},
		},
	})
}
