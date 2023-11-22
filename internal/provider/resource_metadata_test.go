package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceMetadata(t *testing.T) {
	server := newResourceServer(t, "/api/v2/metadata", "1", "password")
	defer server.Close()

	var owner_id = "123"
	var owner_type = "Monitor"

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

				resource "betteruptime_metadata" "this" {
					owner_id   = "%s"
					owner_type = "%s"
					key        = "test"
					value      = "test"
				}
				`, owner_id, owner_type),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_metadata.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_id", owner_id),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_type", owner_type),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "key", "test"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "value", "test"),
				),
			},
			// Step 2 - update.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_metadata" "this" {
					owner_id   = "%s"
					owner_type = "%s"
					key        = "test"
					value      = "test1"
				}
				`, owner_id, owner_type),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_metadata.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_id", owner_id),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "owner_type", owner_type),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "key", "test"),
					resource.TestCheckResourceAttr("betteruptime_metadata.this", "value", "test1"),
				),
			},
			// Step 3 - make no changes, check plan is empty.
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_metadata" "this" {
					owner_id   = "%s"
					owner_type = "%s"
					key        = "test"
					value      = "test1"
				}
				`, owner_id, owner_type),
				PlanOnly: true,
			},
			// Step 4 - destroy.
			{
				ResourceName:      "betteruptime_metadata.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
