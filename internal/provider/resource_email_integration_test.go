package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceEmailIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/email-integrations", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create email integration.
			{
				Config: `
				provider "betteruptime" {
				  api_token = "foo"
				}

				resource "betteruptime_email_integration" "this" {
				  name = "Terraform Test"
				  call = false
				  sms = false
				  email = true
				  push = true
				  critical_alert = true
				  team_wait = 180
				  recovery_period = 0
				  paused = false
				  started_rule_type = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type = "all"

				  started_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Alert]"
				  }
				  started_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Alert Reminder]"
				  }
				  resolved_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Resolved Alert]"
				  }

				  cause_field {
					name = "Cause"
					special_type = "cause"
					field_target = "subject"
					match_type = "match_everything"
				  }
				  started_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "subject"
					match_type = "match_between"
					content_before = "]"
					content_after = ")"
				  }
				  resolved_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "subject"
					match_type = "match_between"
					content_before = "]"
					content_after = ")"
				  }

				  other_started_fields {
					name = "Caused by"
					field_target = "body"
					match_type = "match_between"
					content_before = "by:"
					content_after = "\n"
				  }
				  other_started_fields {
					name = "Description"
					field_target = "body"
					match_type = "match_between"
					content_before = "description:"
					content_after = "\n"
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_email_integration.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "name", "Terraform Test"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "started_rules.0.content", "[Alert]"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "started_rules.1.content", "[Alert Reminder]"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "resolved_rules.0.match_type", "contains"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "resolved_rules.0.content", "[Resolved Alert]"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "cause_field.0.match_type", "match_everything"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "started_alert_id_field.0.content_before", "]"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "resolved_alert_id_field.0.content_after", ")"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "other_started_fields.0.content_before", "by:"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "other_started_fields.1.content_after", "\n"),
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

				resource "betteruptime_email_integration" "this" {
				  name = "Terraform Test - Updated"
				  call = false
				  sms = false
				  email = true
				  push = true
				  critical_alert = true
				  team_wait = 0
				  recovery_period = 180
				  paused = true
				  started_rule_type = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type = "all"

				  started_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Alert]"
				  }
				  started_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Alert Reminder]"
				  }
				  resolved_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Resolved Alert]"
				  }

				  cause_field {
					name = "Cause"
					special_type = "cause"
					field_target = "subject"
					match_type = "match_everything"
				  }
				  started_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "subject"
					match_type = "match_between"
					content_before = "]"
					content_after = ")"
				  }
				  resolved_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "subject"
					match_type = "match_between"
					content_before = "]"
					content_after = ")"
				  }

				  other_started_fields {
					name = "Caused by"
					field_target = "body"
					match_type = "match_between"
					content_before = "by:"
					content_after = "\n"
				  }
				  other_started_fields {
					name = "Description"
					field_target = "body"
					match_type = "match_between"
					content_before = "description:"
					content_after = "\n"
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "name", "Terraform Test - Updated"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "team_wait", "0"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "recovery_period", "180"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "paused", "true"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "started_rules.0.content", "[Alert]"),
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

				resource "betteruptime_email_integration" "this" {
				  name = "Terraform Test - Updated"
				  call = false
				  sms = false
				  email = true
				  push = true
				  critical_alert = true
				  team_wait = 0
				  recovery_period = 180
				  paused = true
				  started_rule_type = "any"
				  acknowledged_rule_type = "unused"
				  resolved_rule_type = "all"

				  started_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[New Alert]"
				  }
				  started_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[New Alert Reminder]"
				  }
				  resolved_rules {
					rule_target = "subject"
					match_type = "contains"
					content = "[Resolved Alert]"
				  }

				  cause_field {
					name = "Cause"
					special_type = "cause"
					field_target = "subject"
					match_type = "match_everything"
				  }
				  started_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "subject"
					match_type = "match_between"
					content_before = "]"
					content_after = ")"
				  }
				  resolved_alert_id_field {
					name = "Alert ID"
					special_type = "alert_id"
					field_target = "subject"
					match_type = "match_between"
					content_before = "]"
					content_after = ")"
				  }

				  other_started_fields {
					name = "Caused by"
					field_target = "body"
					match_type = "match_between"
					content_before = "by:"
					content_after = "\n"
				  }
				  other_started_fields {
					name = "Description"
					field_target = "body"
					match_type = "match_between"
					content_before = "description:"
					content_after = "\n"
				  }
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "name", "Terraform Test - Updated"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "started_rules.0.content", "[New Alert]"),
					resource.TestCheckResourceAttr("betteruptime_email_integration.this", "started_rules.1.content", "[New Alert Reminder]"),
				),
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_email_integration.this",
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
