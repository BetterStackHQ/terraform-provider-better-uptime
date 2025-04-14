package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
		ForceNew:     true,
	},
	"owner_id": {
		Description: "The ID of the owner of this Metadata.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"key": {
		Description: "The key of this Metadata.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
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
		CreateContext: metadataCreateOrUpdate,
		ReadContext:   metadataRead,
		UpdateContext: metadataCreateOrUpdate,
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
	ID        *string          `json:"id,omitempty"`
	OwnerType *string          `json:"owner_type"`
	OwnerID   *string          `json:"owner_id"`
	Key       *string          `json:"key"`
	Values    *[]metadataValue `mapstructure:"metadata_value" json:"values"`
	CreatedAt *string          `json:"created_at,omitempty"`
	UpdatedAt *string          `json:"updated_at,omitempty"`
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
		{k: "metadata_value", v: &in.Values},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func metadataCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in metadata
	for _, e := range metadataRef(&in) {
		if e.k == "metadata_value" {
			loadMetadataValues(d, &in)
		} else if e.k != "id" && e.k != "created_at" && e.k != "updated_at" {
			load(d, e.k, e.v)
		}
	}
	var out metadataHTTPResponse
	if err := metadataPost(ctx, meta, &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return metadataCopyAttrs(d, &out.Data.Attributes)
}

func metadataPost(ctx context.Context, meta interface{}, in *metadata, out *metadataHTTPResponse) diag.Diagnostics {
	reqBody, err := json.Marshal(&in)
	apiUrl := "/api/v3/metadata"
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("POST %s: %s", apiUrl, string(reqBody))
	res, err := meta.(*client).Post(ctx, apiUrl, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		// Keep-Alive.
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	// There are 3 success status codes (200, 201, and 204) in this API endpoint, we don't need to distinguish between them
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return diag.Errorf("POST %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	}
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("POST %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	if res.StatusCode == http.StatusNoContent {
		// Return early on no content
		return nil
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func metadataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out metadataListHTTPResponse
	apiUrl := fmt.Sprintf("/api/v3/metadata?owner_id=%s&owner_type=%s", url.PathEscape(d.Get("owner_id").(string)), url.PathEscape(d.Get("owner_type").(string)))
	if err, ok := resourceRead(ctx, meta, apiUrl, &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404
		return nil
	}

	// Find the matching metadata entry using key
	key := d.Get("key").(string)
	for _, data := range out.Data {
		if data.Attributes.Key != nil && *data.Attributes.Key == key {
			return metadataCopyAttrs(d, &data.Attributes)
		}
	}

	d.SetId("") // Force "create" if no matching key found
	return nil
}

func loadMetadataValues(d *schema.ResourceData, receiver *metadata) {
	metadataStringValue := d.Get("value").(string)
	metadataValues := d.Get("metadata_value").([]interface{})

	// Handle legacy string value
	if metadataStringValue != "" {
		(*receiver).Values = &[]metadataValue{{Type: "String", Value: &metadataStringValue}}
		return
	}

	// Handle metadata_value blocks, ignore any empty values
	if len(metadataValues) > 0 {
		values := make([]metadataValue, 0, len(metadataValues))
		for _, v := range metadataValues {
			valueMap := v.(map[string]interface{})
			value := metadataValue{}

			if v, ok := valueMap["type"].(string); ok && v != "" {
				value.Type = v
			}
			if v, ok := valueMap["value"].(string); ok && v != "" {
				value.Value = &v
			}
			if v, ok := valueMap["item_id"].(string); ok && v != "" {
				value.ItemID = json.Number(v)
			}
			if v, ok := valueMap["email"].(string); ok && v != "" {
				value.Email = &v
			}
			if v, ok := valueMap["name"].(string); ok && v != "" {
				value.Name = &v
			}

			values = append(values, value)
		}
		(*receiver).Values = &values
	}
}

func metadataCopyAttrs(d *schema.ResourceData, in *metadata) diag.Diagnostics {
	var derr diag.Diagnostics

	// Remove item_id, name, and email from metadata values if they were not configured, avoid loading them from API
	if in.Values != nil {
		for valueIndex := range *in.Values {
			if value, ok := d.GetOk(fmt.Sprintf("metadata_value.%d.item_id", valueIndex)); !ok || value == "" {
				(*in.Values)[valueIndex].ItemID = ""
			}
			if value, ok := d.GetOk(fmt.Sprintf("metadata_value.%d.email", valueIndex)); !ok || value == "" {
				(*in.Values)[valueIndex].Email = nil
			}
			if value, ok := d.GetOk(fmt.Sprintf("metadata_value.%d.name", valueIndex)); !ok || value == "" {
				(*in.Values)[valueIndex].Name = nil
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

func metadataDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in metadata
	for _, e := range metadataRef(&in) {
		if e.k != "id" && e.k != "created_at" && e.k != "updated_at" && e.k != "metadata_value" {
			load(d, e.k, e.v)
		}
	}
	return metadataPost(ctx, meta, &in, nil)
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
