package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

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
		Deprecated: "Team name doesn't have to be specified for this resource.",
	},
	"id": {
		Description: "The ID of this Metadata.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"owner_type": {
		Description:  "The type of the owner of this Metadata. Valid values: `Monitor`, `Heartbeat`, `Incident`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`, `CallRouting`",
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"Monitor", "Heartbeat, Incident", "WebhookIntegration", "EmailIntegration", "IncomingWebhook", "CallRouting"}, false),
		Required:     true,
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
		Description: "The value of this Metadata. This field is deprecated, use repeatable block metadata_value to define values with types instead.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     nil,
		Deprecated:  "Use repeatable block metadata_value to define values with types instead.",
	},
	"metadata_value": {
		Description: "An array of typed metadata values of this Metadata.",
		Type:        schema.TypeList,
		Optional:    true,
		Default:     nil,
		Elem:        &schema.Resource{Schema: metadataValueSchema},
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

const scalarMetadataTypesCount = 1

var metadataValueSchema = map[string]*schema.Schema{
	"type": {
		Description: "Value types can be grouped into 2 main categories:\n" +
			"  - **Scalar**: `" + strings.Join(metadataTypes[:scalarMetadataTypesCount], "`, `") + "`\n" +
			"  - **Reference**: `" + strings.Join(metadataTypes[scalarMetadataTypesCount:], "`, `") + "`\n" +
			"  \n" +
			"  The value of a **Scalar** type is defined using the value field.\n" +
			"  \n" +
			"  The value of a **Reference** type is defined using one of the following fields:\n" +
			"  - `item_id` - great choice when you know the ID of the target item.\n" +
			"  - `email` - your go-to choice when you're referencing users.\n" +
			"  - `name` - can be used to reference other items like teams, policies, etc.\n" +
			"  \n" +
			"  **The reference types require the presence of at least one of the three fields: `item_id`, `name`, `email`.**\n",
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
		Description: "ID of the referenced item when type is different than `String`.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"name": {
		Description: "Name of the referenced item when type is different than `String`.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"email": {
		Description: "Email of the referenced user when type is `User`.",
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
		CustomizeDiff: validateMetadata,
		Description:   "https://betterstack.com/docs/uptime/api/metadata/",
		Schema:        metadataSchema,
	}
}

type metadata struct {
	ID             *string          `json:"id,omitempty"`
	OwnerType      *string          `json:"owner_type,omitempty"`
	OwnerID        *string          `json:"owner_id,omitempty"`
	Key            *string          `json:"key,omitempty"`
	MetadataValues *[]metadataValue `mapstructure:"metadata_value" json:"values"`
	CreatedAt      *string          `json:"created_at,omitempty"`
	UpdatedAt      *string          `json:"updated_at,omitempty"`
}

type metadataHTTPResponse struct {
	Data struct {
		ID         string   `json:"id"`
		Attributes metadata `json:"attributes"`
	} `json:"data"`
}

type metadataListHTTPResponse struct {
	Data []struct {
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
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func metadataCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in metadata
	for _, e := range metadataRef(&in) {
		load(d, e.k, e.v)
	}
	var out metadataHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v3/metadata", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return metadataCopyAttrs(d, &out.Data.Attributes)
}

func metadataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out metadataListHTTPResponse
	apiUrl := fmt.Sprintf("/api/v3/metadata/?owner_id=%s&owner_type=%s", url.PathEscape(d.Get("owner_id").(string)), url.PathEscape(d.Get("owner_type").(string)))
	if err, ok := resourceRead(ctx, meta, apiUrl, &out); err != nil {
		return err
	} else if !ok || len(out.Data) == 0 {
		d.SetId("") // Force "create" on 404 or empty list
		return nil
	}
	if len(out.Data) > 1 {
		return diag.Errorf("Found multiple metadata entries for owner_id=%s and owner_type=%s", d.Get("owner_id").(string), d.Get("owner_type").(string))
	}
	return metadataCopyAttrs(d, &out.Data[0].Attributes)
}

func loadMetadataValues(d *schema.ResourceData, receiver **metadataValue) {

	// Convert legacy metadata_values to metadata_value blocks
	if metadataValues, ok := stepValuesObject["metadata_values"].([]interface{}); ok && len(metadataValues) > 0 {
		var metadataValueBlocks []map[string]interface{}
		for _, value := range metadataValues {
			if strValue, ok := value.(string); ok {
				metadataValueBlocks = append(metadataValueBlocks, map[string]interface{}{
					"type":  "String",
					"value": strValue,
				})
			}
		}
		stepValuesObject["metadata_value"] = metadataValueBlocks
		stepValuesObject["metadata_values"] = nil
	}

	// Remove default values ("" or 0) from metadata values
	metadataValues, ok := stepValuesObject["metadata_value"].([]interface{})
	if ok {
		for _, value := range metadataValues {
			valueObject := value.(map[string]interface{})
			for k, v := range valueObject {
				if v == "" || v == 0 {
					valueObject[k] = nil
				}
			}
		}
	}
}

func metadataCopyAttrs(d *schema.ResourceData, in *metadata) diag.Diagnostics {
	var derr diag.Diagnostics

	// Remove item_id, name, and email from metadata values if they were not configured, avoid loading them from API
	if in.MetadataValues != nil {
		for valueIndex := range *in.MetadataValues {
			if value, ok := d.GetOk(fmt.Sprintf("metadata_value.%d.item_id", valueIndex)); !ok || value == "" {
				(*in.MetadataValues)[valueIndex].ItemID = ""
			}
			if value, ok := d.GetOk(fmt.Sprintf("metadata_value.%d.email", valueIndex)); !ok || value == "" {
				(*in.MetadataValues)[valueIndex].Email = nil
			}
			if value, ok := d.GetOk(fmt.Sprintf("metadata_value.%d.name", valueIndex)); !ok || value == "" {
				(*in.MetadataValues)[valueIndex].Name = nil
			}
		}
	}

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
		if e.k != "created_at" && e.k != "updated_at" {
			load(d, e.k, e.v)
		}
	}

	// Using POST requests for updates
	return resourceCreate(ctx, meta, "/api/v3/metadata", &in, &out)
}

func metadataDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, "/api/v3/metadata")
}

func validateMetadata(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	metadataStringValue := d.Get("value").(string)
	metadataValues := d.Get("metadata_value").([]interface{})

	if len(metadataStringValue) == 0 && len(metadataValues) == 0 {
		return fmt.Errorf("there must be at least 1 metadata_value defined, or the deprecated field value must be used")
	}
	if len(metadataStringValue) > 0 && len(metadataValues) > 0 {
		return fmt.Errorf("there cannot be metadata_value defined, when the deprecated field value is used")
	}

	for i, v := range metadataValues {
		value := v.(map[string]interface{})
		if err := validateMetadataValue(value, fmt.Sprintf("metadata_value.%d", i)); err != nil {
			return err
		}
	}
	return nil
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
