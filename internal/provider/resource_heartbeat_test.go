package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceHeartbeat(t *testing.T) {
	server := newResourceServer(t, "/api/v2/heartbeats", "1")
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

				resource "betteruptime_heartbeat" "this" {
					name           = "%s"
					period         = 30
					grace          = 0
					call           = false
					sms            = true
					email          = true
					push           = true
					critical_alert = false
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_heartbeat.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "period", "30"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "grace", "0"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "call", "false"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "sms", "true"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "email", "true"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "push", "true"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "critical_alert", "false"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_heartbeat" "this" {
					name           = "%s"
					period         = 31
					grace          = 1
					call           = true
					sms            = false
					email          = true
					push           = true
					critical_alert = true
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_heartbeat.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "period", "31"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "grace", "1"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "call", "true"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "sms", "false"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "email", "true"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "push", "true"),
					resource.TestCheckResourceAttr("betteruptime_heartbeat.this", "critical_alert", "true"),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_heartbeat" "this" {
					name           = "%s"
					period         = 31
					grace          = 1
					call           = true
					sms            = false
					email          = true
					push           = true
					critical_alert = true
				}
				`, name),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_heartbeat.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
