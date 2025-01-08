package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestJiraIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/jira-integrations", "1")
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

				resource "betteruptime_jira_integration" "this" {
					id 												= "1"
					name                   		= "Test"
					automatic_issue_creation 	= true
					jira_project_key       		= "PROJ"
					jira_issue_type_id     		= "10001"
					jira_fields            		= "{\"customfield_10000\":\"value\"}"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "id", "1"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "name", "Test"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "automatic_issue_creation", "true"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_project_key", "PROJ"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_issue_type_id", "10001"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_fields", "{\"customfield_10000\":\"value\"}"),
				),
			},
			// Step 2 - make no changes, check plan is empty.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_jira_integration" "this" {
					id 												= "1"
					name                   		= "Test"
					automatic_issue_creation 	= true
					jira_project_key       		= "PROJ"
					jira_issue_type_id     		= "10001"
					jira_fields            		= "{\"customfield_10000\":\"value\"}"
				}
				`,
				PlanOnly: true,
			},
			// Step 3 - destroy.
			{
				ResourceName:      "betteruptime_jira_integration.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
