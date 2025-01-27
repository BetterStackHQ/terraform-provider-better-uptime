package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceCatalogRecord(t *testing.T) {
	server := newResourceServer(t, "/api/v2/catalog/relations/123/records", "456")
	defer server.Close()

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
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					
					attribute {
						attribute_id = "789"
						type        = "String"
						value       = "Test Value"
					}

					attribute {
						attribute_id = "790"
						type        = "User"
						email       = "test@example.com"
					}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_catalog_record.test", "relation_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_catalog_record.test", "attribute.#", "2"),
				),
			},
			// Step 2 - update
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					
					attribute {
						attribute_id = "789"
						type        = "String"
						value       = "Updated Value"
					}

					attribute {
						attribute_id = "790"
						type        = "User"
						email       = "updated@example.com"
					}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_catalog_record.test", "relation_id", "123"),
					resource.TestCheckResourceAttr("betteruptime_catalog_record.test", "attribute.#", "2"),
				),
			},
			// Step 3 - import
			{
				ResourceName:      "betteruptime_catalog_record.test",
				ImportState:       true,
				ImportStateId:     "123/456",
				ImportStateVerify: true,
			},
		},
	})
}
