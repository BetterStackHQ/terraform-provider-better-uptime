package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceCatalogRecord(t *testing.T) {
	server := newResourceServer(t, "/api/v2/catalog/relations/123/records", "2")
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
		},
	})
}

func TestResourceCatalogRecordValidation(t *testing.T) {
	server := newResourceServer(t, "/api/v2/catalog/relations/123/records", "2")
	defer server.Close()

	cases := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "valid string attribute",
			config: `
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
				}
			`,
			expectError: nil,
		},
		{
			name: "invalid string attribute - missing value",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					attribute {
						attribute_id = "789"
						type        = "String"
					}
				}
			`,
			expectError: regexp.MustCompile("value must be set for String type attribute"),
		},
		{
			name: "invalid string attribute - with item_id",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					attribute {
						attribute_id = "789"
						type        = "String"
						value       = "Test"
						item_id     = "123"
					}
				}
			`,
			expectError: regexp.MustCompile("item_id must not be set for String type attribute"),
		},
		{
			name: "valid user attribute",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					attribute {
						attribute_id = "789"
						type        = "User"
						email       = "test@example.com"
					}
				}
			`,
			expectError: nil,
		},
		{
			name: "invalid user attribute - with value",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					attribute {
						attribute_id = "789"
						type        = "User"
						email       = "test@example.com"
						value       = "Test"
					}
				}
			`,
			expectError: regexp.MustCompile("value must not be set for User type attribute"),
		},
		{
			name: "invalid user attribute - no identifier",
			config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_catalog_record" "test" {
					relation_id = "123"
					attribute {
						attribute_id = "789"
						type        = "User"
					}
				}
			`,
			expectError: regexp.MustCompile("at least one of item_id, email, or name must be set for User type attribute"),
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
