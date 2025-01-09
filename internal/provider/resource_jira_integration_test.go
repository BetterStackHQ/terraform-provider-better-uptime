package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestJiraIntegration(t *testing.T) {
	server := newResourceServer(t, "/api/v2/jira-integrations", "42")
	defer server.Close()
	// Initialize the resource, since it's not created by a POST request
	server.Data.Store([]byte(`{
		"name": "Original name",
		"automatic_issue_creation": false,
		"jira_project_key": "PROJ",
		"jira_issue_type_id": "10001",
		"jira_fields": "{}"
	}`))

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
					better_stack_id          = "42"
					name                     = "Test"
					automatic_issue_creation = true
					jira_project_key         = "PROJ"
					jira_issue_type_id       = "10001"
					jira_fields              = "{\"customfield_10000\":\"value\"}"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "id", "42"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "name", "Test"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "automatic_issue_creation", "true"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_project_key", "PROJ"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_issue_type_id", "10001"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_fields", "{\"customfield_10000\":\"value\"}"),
				),
			},
			// Step 2 - change only automatic_issue_creation, omit the rest
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_jira_integration" "this" {
					better_stack_id          = "42"
					automatic_issue_creation = false
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "id", "42"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "name", "Test"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "automatic_issue_creation", "false"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_project_key", "PROJ"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_issue_type_id", "10001"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_fields", "{\"customfield_10000\":\"value\"}"),
				),
			},
		},
	})
}
