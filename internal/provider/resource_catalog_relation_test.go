package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceCatalogRelation(t *testing.T) {
	server := newResourceServer(t, "/api/v2/catalog/relations", "1")
	defer server.Close()

	var name = "Office"
	var description = "A physical office building representing ACME Group"
	var updatedDescription = "Updated description for office building"

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

				resource "betteruptime_catalog_relation" "this" {
					name        = "%s"
					description = "%s"
				}
				`, name, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_catalog_relation.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_catalog_relation.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_catalog_relation.this", "description", description),
				),
			},
			// Step 2 - update
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_relation" "this" {
					name        = "%s"
					description = "%s"
				}
				`, name, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_catalog_relation.this", "id"),
					resource.TestCheckResourceAttr("betteruptime_catalog_relation.this", "name", name),
					resource.TestCheckResourceAttr("betteruptime_catalog_relation.this", "description", updatedDescription),
				),
			},
			// Step 3 - make no changes, check plan is empty
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_relation" "this" {
					name        = "%s"
					description = "%s"
				}
				`, name, updatedDescription),
				PlanOnly: true,
			},
			// Step 4 - import
			{
				ResourceName:      "betteruptime_catalog_relation.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
