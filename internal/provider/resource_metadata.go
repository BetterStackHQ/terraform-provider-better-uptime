package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var metadataSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     nil,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of this Metadata.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"owner_type": {
		Description: "The type of the owner of this Metadata. Valid values: `Monitor`, `Heartbeat`, `Incident`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`",
		Type:        schema.TypeString,
		Required:    true,
	},
	"owner_id": {
		Description: "The ID of the owner of this Metadata.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"key": {
		Description: "The key of this Metadata.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"value": {
		Description: "The value of this Metadata.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"created_at": {
		Description: "The time when this metadata was created.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this metadata was updated.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

var metadataTypes = []string{
	"String", "User", "Team", "Policy", "Schedule",
	"SlackIntegration", "LinearIntegration", "JiraIntegration",
	"MicrosoftTeamsWebhook", "ZapierWebhook", "NativeWebhook",
	"PagerDutyWebhook",
}

var metadataValueSchema = map[string]*schema.Schema{
	"type": {
		Description:  "Type of the value. When left empty, the String type is used.",
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "String",
		ValidateFunc: validation.StringInSlice(metadataTypes, false),
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
	},
	"name": {
		Description: "Human readable name of the referenced item when type is different than String and the item has a name.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"email": {
		Description: "Email of the referenced user when type is User.",
		Type:        schema.TypeString,
		Optional:    true,
	},
}

type metadataValue struct {
	Type   string      `mapstructure:"type" json:"type"`
	Value  *string     `mapstructure:"value,omitempty" json:"value,omitempty"`
	ItemID json.Number `mapstructure:"item_id,omitempty" json:"item_id,omitempty"`
	Name   *string     `mapstructure:"name,omitempty" json:"name,omitempty"`
	Email  *string     `mapstructure:"email,omitempty" json:"email,omitempty"`
}

func newMetadataResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: metadataCreate,
		ReadContext:   metadataRead,
		UpdateContext: metadataUpdate,
		DeleteContext: metadataDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateRequestHeaders,
		Description:   "https://betterstack.com/docs/uptime/api/metadata/",
		Schema:        metadataSchema,
	}
}

type metadata struct {
	ID        *string `json:"id,omitempty"`
	OwnerType *string `json:"owner_type,omitempty"`
	OwnerID   *string `json:"owner_id,omitempty"`
	Key       *string `json:"key,omitempty"`
	Value     *string `json:"value,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	TeamName  *string `json:"team_name,omitempty"`
}

type metadataHTTPResponse struct {
	Data struct {
		ID         string   `json:"id"`
		Attributes metadata `json:"attributes"`
	} `json:"data"`
}

func metadataRef(in *metadata) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "id", v: &in.ID},
		{k: "owner_type", v: &in.OwnerType},
		{k: "owner_id", v: &in.OwnerID},
		{k: "key", v: &in.Key},
		{k: "value", v: &in.Value},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func metadataCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in metadata
	for _, e := range metadataRef(&in) {
		if e.k == "request_headers" {
			if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
				return diag.FromErr(err)
			}
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
	var out metadataHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/metadata", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return metadataCopyAttrs(d, &out.Data.Attributes)
}

func metadataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out metadataHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/metadata/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return metadataCopyAttrs(d, &out.Data.Attributes)
}

func metadataCopyAttrs(d *schema.ResourceData, in *metadata) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range metadataRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func metadataUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in metadata
	var out metadataHTTPResponse
	for _, e := range metadataRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
					return diag.FromErr(err)
				}
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/metadata/%s", url.PathEscape(d.Id())), &in, &out)
}

func metadataDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/metadata/%s", url.PathEscape(d.Id())))
}

func validateMetadataValue(attrMap map[string]interface{}, path string) error {
	attrType := attrMap["type"].(string)

	// Validation for String type
	if attrType == "String" {
		if value, ok := attrMap["value"].(string); !ok || value == "" {
			return fmt.Errorf("%s: value must be set for String type", path)
		}
		if itemID, ok := attrMap["item_id"].(string); ok && itemID != "" {
			return fmt.Errorf("%s: item_id must not be set for String type", path)
		}
		if email, ok := attrMap["email"].(string); ok && email != "" {
			return fmt.Errorf("%s: email must not be set for String type", path)
		}
		if name, ok := attrMap["name"].(string); ok && name != "" {
			return fmt.Errorf("%s: name must not be set for String type", path)
		}
		return nil
	}

	// Validation for non-String types
	if value, ok := attrMap["value"].(string); ok && value != "" {
		return fmt.Errorf("%s: value must not be set for %s type", path, attrType)
	}

	hasIdentifier := false
	if itemID, ok := attrMap["item_id"].(string); ok && itemID != "" {
		hasIdentifier = true
	}
	if email, ok := attrMap["email"].(string); ok && email != "" {
		hasIdentifier = true
	}
	if name, ok := attrMap["name"].(string); ok && name != "" {
		hasIdentifier = true
	}
	if !hasIdentifier {
		return fmt.Errorf("%s: at least one of item_id, email, or name must be set for %s type", path, attrType)
	}

	return nil
}
