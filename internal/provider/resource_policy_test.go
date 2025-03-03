package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourcePolicy(t *testing.T) {
	server := newResourceServer(t, "/api/v3/policies", "1")
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
                    step_members {
                      type = "incident_metadata"
                      metadata_key = "Team"
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
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.step_members.1.type", "incident_metadata"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.step_members.1.metadata_key", "Team"),
				),
				PreConfig: func() {
					t.Log("step 1")
				},
			},
			// Step 2 - change to a branching policy, use legacy metadata_values
			{
				ExpectNonEmptyPlan: true, // Ignoring plan not empty error since we're using legacy metadata_values attribute
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
                    policy_metadata_key = "Team Policy"
                  }
                  steps {
                    type            = "metadata_branching"
                    wait_before     = 0
                    metadata_key    = "severity"
                    metadata_values = ["critical", "error"]
                    policy_id = 456
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
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.policy_metadata_key", "Team Policy"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.type", "metadata_branching"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.wait_before", "0"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_key", "severity"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_value.0.type", "String"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_value.0.value", "critical"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_value.1.type", "String"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.metadata_value.1.value", "error"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.2.policy_id", "456"),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 3 - make no changes, only update to metadata_value blocks, check plan is empty.
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
                    policy_metadata_key = "Team Policy"
                  }
                  steps {
                    type            = "metadata_branching"
                    wait_before     = 0
                    metadata_key    = "severity"
                    metadata_value {
                      value = "critical"
					}
                    metadata_value {
                      value = "error"
					}
                    policy_id = 456
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

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Test time_branching days validation.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name         = "Terraform - Test"

                  steps {
                    type        = "time_branching"
					wait_before = 0
                    timezone    = "Prague"
                    days        = ["sat", "sun", "invalid"]
                    time_from   = "08:00"
                    time_to     = "22:00"
                  }
				}
				`,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: regexp.MustCompile(`expected steps\.0\.days\.2 to be one of \["mon" "tue" "wed" "thu" "fri" "sat" "sun"], got invalid`),
				PreConfig: func() {
					t.Log("test validation: days")
				},
			},
			// Test time_branching time_from validation.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name         = "Terraform - Test"

                  steps {
                    type        = "time_branching"
					wait_before = 0
                    timezone    = "Prague"
                    days        = ["sat", "sun"]
                    time_from   = "8 AM"
                    time_to     = "22:00"
                  }
				}
				`,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: regexp.MustCompile(`invalid value for steps\.0\.time_from \(use HH:MM format\)`),
				PreConfig: func() {
					t.Log("test validation: time_from")
				},
			},
			// Test time_branching time_to validation.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name         = "Terraform - Test"

                  steps {
                    type        = "time_branching"
					wait_before = 0
                    timezone    = "Prague"
                    days        = ["sat", "sun"]
                    time_from   = "08:00"
                    time_to     = "10 PM"
                  }
				}
				`,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: regexp.MustCompile(`invalid value for steps\.0\.time_to \(use HH:MM format\)`),
				PreConfig: func() {
					t.Log("test validation: time_to")
				},
			},
		},
	})

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Test wait_until_time / wait_until_timezone
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name		   = "Terraform - Test"
				  repeat_count = 3
				  repeat_delay = 60

				  steps {
					type		= "escalation"
					wait_before = 0
					urgency_id	= 123
					step_members { type = "current_on_call" }
					step_members {
					  type = "slack_integration"
					  id = 123
					}
				  }
				  steps {
					type				= "escalation"
					wait_until_time		= "07:45"
					wait_until_timezone = "UTC"
					urgency_id			= 123
					step_members {
					  type = "entire_team"
					  id = 123
					}
				  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_policy.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.0.wait_before", "0"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.wait_until_time", "07:45"),
					resource.TestCheckResourceAttr("betteruptime_policy.this", "steps.1.wait_until_timezone", "UTC"),
				),
				PreConfig: func() {
					t.Log("test valid wait_until_time")
				},
			},
		},
	})

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Test wait_until_time validation
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "this" {
				  name		   = "Terraform - Test"
				  repeat_count = 3
				  repeat_delay = 60

				  steps {
					type		= "escalation"
					wait_before = 0
					urgency_id	= 123
					step_members { type = "current_on_call" }
					step_members {
						type = "slack_integration"
						id = 123
					}
				  }
				  steps {
					type				= "escalation"
					wait_until_time		= "9 AM"
					wait_until_timezone = "UTC"
					urgency_id			= 123
					step_members {
					  type = "entire_team"
					  id = 123
					}
				  }
				}
				`,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: regexp.MustCompile(`invalid value for steps\.1\.wait_until_time \(use HH:MM format\)`),
				PreConfig: func() {
					t.Log("test validation: wait_until_time")
				},
			},
		},
	})
}
func TestResourcePolicyMetadataValidation(t *testing.T) {
	server := newResourceServer(t, "/api/v3/policies", "1")
	defer server.Close()

	cases := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "invalid - metadata_value on non-metadata step",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "escalation"
						metadata_value {
							type = "String"
							value = "test"
						}
					}
				}
			`,
			expectError: regexp.MustCompile(`steps\.0: metadata_value must be empty for non-metadata_branching steps`),
		},
		{
			name: "invalid - no metadata_key on metadata step",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_value {
							value = "test"
						}
					}
				}
			`,
			expectError: regexp.MustCompile(`steps\.0: missing metadata_key for metadata_branching step`),
		},
		{
			name: "invalid - no metadata_value on metadata step",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "environment"
					}
				}
			`,
			expectError: regexp.MustCompile(`steps\.0: there must be at least 1 metadata_value for metadata_branching step`),
		},
		{
			name: "invalid - metadata value missing value",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "environment"
						metadata_value {
							type = "String"
						}
					}
				}
			`,
			expectError: regexp.MustCompile(`steps\.0\.metadata_value\.0: value must be set for String type`),
		},
		{
			name: "invalid - metadata value in non-String",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "environment"
						metadata_value {
							type = "User"
							value = "My user"
							email = "user@email.com"
						}
					}
				}
			`,
			expectError: regexp.MustCompile(`steps\.0\.metadata_value\.0: value must not be set for User type`),
		},
		{
			name: "invalid - no identification in non-String",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "environment"
						metadata_value {
							type = "User"
						}
					}
				}
			`,
			expectError: regexp.MustCompile(`steps\.0\.metadata_value\.0: at least one of item_id, email, or name must be set for User type`),
		},
		{
			name: "valid - metadata branching with values",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "environment"
						metadata_value {
							type = "String"
							value = "production"
						}
						metadata_value {
							type = "String"
							value = "staging"
						}
						policy_id = 123
					}
				}
			`,
			expectError: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"betteruptime": func() (*schema.Provider, error) {
						return New(WithURL(server.URL)), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

func TestResourcePolicyMetadataValueStateCleanup(t *testing.T) {
	server := newResourceServer(t, "/api/v3/policies", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - Create with User type
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "owner"
						metadata_value {
							type = "User"
							email = "test@example.com"
						}
					}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					// Simulate API setting computed fields
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["betteruptime_policy.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						rs.Primary.Attributes["steps.0.metadata_value.0.item_id"] = "456"
						rs.Primary.Attributes["steps.0.metadata_value.0.name"] = "Test User"
						return nil
					},
					// Verify all User type fields are present
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.type", "User"),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.email", "test@example.com"),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.item_id", "456"),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.name", "Test User"),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.value", ""),
				),
			},
			// Step 2 - Update to String type, should clean up computed fields
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_policy" "test" {
					name = "Test Policy"
					steps {
						type = "metadata_branching"
						metadata_key = "owner"
						metadata_value {
							type = "String"
							value = "test@example.com"
						}
					}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					// Verify only String type fields are present
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.type", "String"),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.value", "test@example.com"),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.email", ""),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.item_id", ""),
					resource.TestCheckResourceAttr("betteruptime_policy.test", "steps.0.metadata_value.0.name", ""),
				),
			},
		},
	})
}
