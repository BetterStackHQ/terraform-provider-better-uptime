package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourcePolicy(t *testing.T) {
	server := newResourceServer(t, "/api/v2/policies", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create an escalation policy.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name         = "Terraform - Test"
				  repeat_count = 3
				  repeat_delay = 60

				  steps {
					type        = "escalation"
					wait_before = 0
					urgency_id  = 123
					step_members { type = "current_on_call" }
					step_members {
                      type = "slack_integration"
                      id = 123
                    }
				  }
				  steps {
					type        = "escalation"
					wait_before = 180
					urgency_id  = 123
					step_members {
                      type = "entire_team"
                      id = 123
                    }
				  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_policy.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "name", "Terraform - Test"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.urgency_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.urgency_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.step_members.0.type", "current_on_call"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.step_members.1.type", "slack_integration"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.step_members.1.id", "123"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.step_members.0.type", "entire_team"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.step_members.0.id", "123"),
				),
				PreConfig: func() {
					t.Log("step 1")
				},
			},
			// Step 2 - change to a branching policy.
			{
				Config: `
                provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name = "Terraform - Branching"
                  steps {
                    type        = "time_branching"
                    wait_before = 0
                    timezone    = "Prague"
                    days        = ["mon", "tue", "wed", "thu", "fri"]
                    time_from   = "08:00"
                    time_to     = "22:00"
                    policy_id   = 456
                  }
                  steps {
                    type        = "time_branching"
                    wait_before = 0
                    timezone    = "Prague"
                    days        = ["sat", "sun"]
                    time_from   = "08:00"
                    time_to     = "22:00"
                    policy_id   = 456
                  }
                  steps {
                    type            = "metadata_branching"
                    wait_before     = 0
                    metadata_key    = "severity"
                    metadata_values = ["critical", "error"]
                    policy_id       = 456
                  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_policy.this", "name", "Terraform - Branching"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.type", "time_branching"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.wait_before", "0"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.timezone", "Prague"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.days.0", "mon"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.time_from", "08:00"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.time_to", "22:00"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.policy_id", "456"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.type", "time_branching"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.wait_before", "0"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.timezone", "Prague"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.days.0", "sat"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.time_to", "22:00"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.policy_id", "456"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.type", "metadata_branching"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.wait_before", "0"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_key", "severity"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_values.0", "critical"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_values.1", "error"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.policy_id", "456"),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: `
                provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name = "Terraform - Branching"
                  steps {
                    type        = "time_branching"
                    wait_before = 0
                    timezone    = "Prague"
                    days        = ["mon", "tue", "wed", "thu", "fri"]
                    time_from   = "08:00"
                    time_to     = "22:00"
                    policy_id   = 456
                  }
                  steps {
                    type        = "time_branching"
                    wait_before = 0
                    timezone    = "Prague"
                    days        = ["sat", "sun"]
                    time_from   = "08:00"
                    time_to     = "22:00"
                    policy_id   = 456
                  }
                  steps {
                    type            = "metadata_branching"
                    wait_before     = 0
                    metadata_key    = "severity"
                    metadata_values = ["critical", "error"]
                    policy_id       = 456
                  }
				}`,
				PlanOnly: true,
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_policy.this",
				ImportState:       true,
				ImportStateId:     "1",
				ImportStateVerify: true,
				PreConfig: func() {
					t.Log("step 4")
				},
			},
		},
	})
}
