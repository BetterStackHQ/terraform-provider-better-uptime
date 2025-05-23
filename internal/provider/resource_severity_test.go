package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceSeverity(t *testing.T) {
	server := newResourceServer(t, "/api/v2/urgencies", "1")
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create a severity.
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity" "this" {
				  name           = "Terraform - Test"
				  sms            = true
				  call           = false
				  email          = false
				  push           = true
				  critical_alert = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_severity.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "name", "Terraform - Test"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "sms", "true"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "call", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "email", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "push", "true"),
				),
				PreConfig: func() {
					t.Log("step 1")
				},
			},
			// Step 2 - change to a call-only severity.
			{
				Config: `
                provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity" "this" {
				  name           = "Terraform - Call only"
				  sms            = false
				  call           = true
				  email          = false
				  push           = false
				  critical_alert = false
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_severity.this", "name", "Terraform - Call only"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "sms", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "call", "true"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "email", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "push", "false"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/urgencies/1", `{"name":"Terraform - Call only","sms":false,"call":true,"email":false,"push":false,"critical_alert":false}`),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 3 - set severity group.
			{
				Config: `
                provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity" "this" {
				  name              = "Terraform - Call only"
				  sms               = false
				  call              = true
				  email             = false
				  push              = false
				  critical_alert    = false
				  severity_group_id = 123
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_severity.this", "name", "Terraform - Call only"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "sms", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "call", "true"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "email", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "push", "false"),
					resource.TestCheckResourceAttr("betteruptime_severity.this", "severity_group_id", "123"),
					server.TestCheckCalledRequest("PATCH", "/api/v2/urgencies/1", `{"name":"Terraform - Call only","sms":false,"call":true,"email":false,"push":false,"critical_alert":false,"urgency_group_id":123}`),
				),
				PreConfig: func() {
					t.Log("step 2")
				},
			},
			// Step 4 - make no changes, check plan is empty.
			{
				Config: `
                provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity" "this" {
				  name              = "Terraform - Call only"
				  sms               = false
				  call              = true
				  email             = false
				  push              = false
				  critical_alert    = false
				  severity_group_id = 123
				}`,
				PlanOnly: true,
				PreConfig: func() {
					t.Log("step 3")
				},
			},
			// Step 5 - destroy.
			{
				ResourceName:      "betteruptime_severity.this",
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
