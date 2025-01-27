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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var catalogRecordSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Catalog Record.",
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
		Description: "List of attribute values for the Catalog record.",
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
				"type": {
					Description: "Type of the value.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: validation.StringInSlice([]string{
						"String", "User", "Team", "Policy", "Schedule",
						"SlackIntegration", "LinearIntegration", "JiraIntegration",
						"MicrosoftTeamsWebhook", "ZapierWebhook", "NativeWebhook",
						"PagerDutyWebhook",
					}, false),
				},
				"value": {
					Description: "Value when type is String.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"item_id": {
					Description: "ID of the referenced item when type is different than String.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
				"name": {
					Description: "Human readable name of the referenced item when type is different than String and the item has a name.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
				"email": {
					Description: "Email of the referenced user when type is User.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
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
		Description: "https://betterstack.com/docs/uptime/api/create-catalog-record/",
		Schema:      catalogRecordSchema,
	}
}

type catalogRecordValue struct {
	Type   string      `json:"type"`
	Value  *string     `json:"value,omitempty"`
	ItemID json.Number `json:"item_id,omitempty"`
	Name   *string     `json:"name,omitempty"`
	Email  *string     `json:"email,omitempty"`
}

type catalogRecordAttribute struct {
	Attribute struct {
		ID   json.Number `json:"id"`
		Name string      `json:"name"`
	} `json:"attribute"`
	Values []catalogRecordValue `json:"values"`
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

		var value catalogRecordValue
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

		attribute.Values = []catalogRecordValue{value}
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
	if err := d.Set("attribute", flattenCatalogRecordAttributes(in.Attributes)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
