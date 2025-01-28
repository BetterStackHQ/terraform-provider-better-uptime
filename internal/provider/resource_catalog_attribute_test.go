package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceCatalogAttribute(t *testing.T) {
	server := newResourceServer(t, "/api/v2/catalog/relations/123/attributes", "2")
	defer server.Close()

	var name = "Address"
	var updatedName = "Location"
	var primary = true

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_attribute" "this" {
					relation_id = "123"
					name       = "%s"
					primary    = %t
				}
				`, name, primary),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_catalog_attribute.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_catalog_attribute.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_catalog_attribute.this", "primary", fmt.Sprintf("%t", primary)),
					resource.TestCheckResourceAttr("betteruptime_catalog_attribute.this", "relation_id", "123"),
				),
			},
			// Step 2 - update
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_attribute" "this" {
					relation_id = "123"
					name       = "%s"
					primary    = %t
				}
				`, updatedName, primary),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_catalog_attribute.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_catalog_attribute.this", "name", updatedName),
					resource.TestCheckResourceAttr("betteruptime_catalog_attribute.this", "primary", fmt.Sprintf("%t", primary)),
				),
			},
			// Step 3 - make no changes, check plan is empty
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_attribute" "this" {
					relation_id = "123"
					name       = "%s"
					primary    = %t
				}
				`, updatedName, primary),
				PlanOnly: true,
			},
			// Step 4 - import
			{
				ResourceName:      "betteruptime_catalog_attribute.this",
				ImportState:       true,
				ImportStateId:     "123/2",
				ImportStateVerify: true,
			},
		},
	})
}
