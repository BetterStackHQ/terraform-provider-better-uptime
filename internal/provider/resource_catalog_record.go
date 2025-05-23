package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var catalogRecordSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Catalog record.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"relation_id": {
		Description: "The ID of the Catalog relation this record belongs to.",
		Type:        schema.TypeString,
		ForceNew:    true,
		Required:    true,
	},
	"attribute": {
		Description: "List of attribute values for the Catalog record. You can have multiple blocks with same `attribute_id` for multiple values.",
		Type:        schema.TypeList,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"attribute_id": {
					Description: "ID of the target Catalog attribute.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"attribute_name": {
					Description: "Name of the Catalog attribute.",
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    false,
				},
				"type":    metadataValueSchema["type"],
				"value":   metadataValueSchema["value"],
				"item_id": metadataValueSchema["item_id"],
				"name":    metadataValueSchema["name"],
				"email":   metadataValueSchema["email"],
			},
		},
	},
}

func newCatalogRecordResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: catalogRecordCreate,
		ReadContext:   catalogRecordRead,
		UpdateContext: catalogRecordUpdate,
		DeleteContext: catalogRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				split := strings.SplitN(d.Id(), "/", 2)
				if len(split) != 2 {
					return nil, errors.New("betteruptime_catalog_record can be imported via \"relation_id/id\" only (e.g. \"123/234\")")
				}
				if err := d.Set("relation_id", split[0]); err != nil {
					return nil, err
				}
				d.SetId(split[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		CustomizeDiff: validateCatalogRecordAttributes,
		Description:   "https://betterstack.com/docs/uptime/api/catalog-integrations-records/",
		Schema:        catalogRecordSchema,
	}
}

type catalogRecordAttribute struct {
	Attribute struct {
		ID   json.Number `json:"id"`
		Name string      `json:"name,omitempty"`
	} `json:"attribute"`
	Values []metadataValue `json:"values"`
}

type catalogRecord struct {
	Attributes []catalogRecordAttribute `json:"attributes"`
}

type catalogRecordHTTPResponse struct {
	Data struct {
		ID         string        `json:"id"`
		Attributes catalogRecord `json:"attributes"`
	} `json:"data"`
}

func expandCatalogRecordAttributes(d *schema.ResourceData) []catalogRecordAttribute {
	attributeList := d.Get("attribute").([]interface{})
	attributes := make([]catalogRecordAttribute, 0, len(attributeList))

	for _, attr := range attributeList {
		attrMap := attr.(map[string]interface{})
		var attribute catalogRecordAttribute
		attribute.Attribute.ID = json.Number(attrMap["attribute_id"].(string))

		var value metadataValue
		value.Type = attrMap["type"].(string)

		if v, ok := attrMap["value"].(string); ok && v != "" {
			value.Value = &v
		}
		if v, ok := attrMap["item_id"].(string); ok && v != "" {
			value.ItemID = json.Number(v)
		}
		if v, ok := attrMap["name"].(string); ok && v != "" {
			value.Name = &v
		}
		if v, ok := attrMap["email"].(string); ok && v != "" {
			value.Email = &v
		}

		attribute.Values = []metadataValue{value}
		attributes = append(attributes, attribute)
	}

	return attributes
}

func flattenCatalogRecordAttributes(attributes []catalogRecordAttribute) []interface{} {
	result := make([]interface{}, 0, len(attributes))

	for _, attr := range attributes {
		for _, value := range attr.Values {
			m := map[string]interface{}{
				"attribute_id":   attr.Attribute.ID.String(),
				"attribute_name": attr.Attribute.Name,
				"type":           value.Type,
			}

			if value.Value != nil {
				m["value"] = *value.Value
			}
			if len(value.ItemID) > 0 {
				m["item_id"] = value.ItemID.String()
			}
			if value.Name != nil {
				m["name"] = *value.Name
			}
			if value.Email != nil {
				m["email"] = *value.Email
			}

			result = append(result, m)
		}
	}

	return result
}

func catalogRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relationID := d.Get("relation_id").(string)

	in := catalogRecord{
		Attributes: expandCatalogRecordAttributes(d),
	}

	var out catalogRecordHTTPResponse
	if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/records", url.PathEscape(relationID)), &in, &out); err != nil {
		return err
	}

	d.SetId(out.Data.ID)
	return catalogRecordCopyAttrs(d, &out.Data.Attributes)
}

func catalogRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relationID := d.Get("relation_id").(string)
	var out catalogRecordHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/records/%s", url.PathEscape(relationID), url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("")
		return nil
	}
	return catalogRecordCopyAttrs(d, &out.Data.Attributes)
}

func catalogRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relationID := d.Get("relation_id").(string)

	in := catalogRecord{
		Attributes: expandCatalogRecordAttributes(d),
	}

	var out catalogRecordHTTPResponse
	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/records/%s", url.PathEscape(relationID), url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	return catalogRecordCopyAttrs(d, &out.Data.Attributes)
}

func catalogRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relationID := d.Get("relation_id").(string)
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/records/%s", url.PathEscape(relationID), url.PathEscape(d.Id())))
}

func catalogRecordCopyAttrs(d *schema.ResourceData, in *catalogRecord) diag.Diagnostics {
	attributes := flattenCatalogRecordAttributes(in.Attributes)

	// Remove item_id, name, and email from attributes if they were not configured, avoid loading them from API
	for attributeIndex, attribute := range attributes {
		attributeMap := attribute.(map[string]interface{})
		for _, field := range []string{"item_id", "name", "email"} {
			if value, ok := d.GetOk(fmt.Sprintf("attribute.%d.%s", attributeIndex, field)); !ok || value == "" {
				delete(attributeMap, field)
			}
		}
	}

	if err := d.Set("attribute", attributes); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func validateCatalogRecordAttributes(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	attributes := d.Get("attribute").([]interface{})

	for i, attr := range attributes {
		attrMap := attr.(map[string]interface{})
		if err := validateMetadataValue(attrMap, fmt.Sprintf("attribute.%d", i)); err != nil {
			return err
		}
	}

	return nil
}
