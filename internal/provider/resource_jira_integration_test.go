package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

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
		"jira_fields": {
			"duedate": 1209600,
			"customfield_10000": "original value",
			"customfield_10001": "to be removed"
		}
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
					jira_fields_json         = jsonencode({
						duedate = 1209600
						customfield_10000 = "new value"
					})
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "id", "42"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "name", "Test"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "automatic_issue_creation", "true"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_project_key", "PROJ"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_issue_type_id", "10001"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_fields_json", "{\"customfield_10000\":\"new value\",\"duedate\":1209600}"),
					func(s *terraform.State) error {
						var lastPatch CalledRequest
						for _, req := range server.CalledRequests {
							if req.Method == "PATCH" {
								lastPatch = req
							}
						}
						// Check that PATCH request JSON contains the numerical value as number
						expectedSubstring := `"jira_fields":{"customfield_10000":"new value","duedate":1209600}`
						if !strings.Contains(lastPatch.Body, expectedSubstring) {
							return fmt.Errorf("expected last PATCH body to contain %s, got %s", expectedSubstring, lastPatch.Body)
						}
						// Check that PATCH request does not contain the JSON field
						if strings.Contains(strings.ToLower(lastPatch.Body), "jirafieldsjson") || strings.Contains(strings.ToLower(lastPatch.Body), "jira_fields_json") {
							return fmt.Errorf("did not expect last PATCH body to contain mention of the JSON field, got %s", lastPatch.Body)
						}
						return nil
					},
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
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_fields_json", "{\"customfield_10000\":\"new value\",\"duedate\":1209600}"),
					func(s *terraform.State) error {
						var lastPatch CalledRequest
						for _, req := range server.CalledRequests {
							if req.Method == "PATCH" {
								lastPatch = req
							}
						}
						// Check that PATCH request JSON does not contain jira_fields
						unexpectedSubstring := `"jira_fields"`
						if strings.Contains(lastPatch.Body, unexpectedSubstring) {
							return fmt.Errorf("did not expect last PATCH body to contain %s, got %s", unexpectedSubstring, lastPatch.Body)
						}
						return nil
					},
				),
			},
			// Step 3 - remove jira_fields
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_jira_integration" "this" {
					better_stack_id          = "42"
					jira_fields_json         = jsonencode({})
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "id", "42"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "name", "Test"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "automatic_issue_creation", "false"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_project_key", "PROJ"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_issue_type_id", "10001"),
					resource.TestCheckResourceAttr("betteruptime_jira_integration.this", "jira_fields_json", "{}"),
					func(s *terraform.State) error {
						var lastPatch CalledRequest
						for _, req := range server.CalledRequests {
							if req.Method == "PATCH" {
								lastPatch = req
							}
						}
						// Check that PATCH request JSON contains the empty jira_fields
						expectedSubstring := `"jira_fields":{}`
						if !strings.Contains(lastPatch.Body, expectedSubstring) {
							return fmt.Errorf("expected last PATCH body to contain %s, got %s", expectedSubstring, lastPatch.Body)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestJiraIntegrationWithoutId(t *testing.T) {
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
					name                     = "Test"
					automatic_issue_creation = true
					jira_project_key         = "PROJ"
					jira_issue_type_id       = "10001"
					jira_fields_json         = jsonencode({
						duedate = 1209600
						customfield_10000 = "new value"
					})
				}
				`,
				ExpectError: regexp.MustCompile(`Due to required authentication in Jira, the integration has to be created and removed in Better Stack web UI\. You can either import the resource, or set the ID of the Jira Integration in better_stack_id field and it will be auto-imported during resource creation\.`),
			},
		},
	})
}
