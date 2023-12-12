package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var metadataSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Metadata.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"owner_type": {
		Description: "The type of the owner of this Metadata. Valid values: `Monitor`, `Heartbeat`, `Incident`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`",
		Type:        schema.TypeString,
		Required:    true,
	},
	"owner_id": {
		Description: "The ID of the owner of this Metadata.",
		Type:        schema.TypeInt,
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
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this metadata was updated.",
		Type:        schema.TypeString,
		Computed:    true,
	},
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
		Description: "https://docs.betteruptime.com/api/metadata-api",
		Schema:      metadataSchema,
	}
}

type metadata struct {
	ID        *int    `json:"id,omitempty"`
	OwnerType *string `json:"owner_type,omitempty"`
	OwnerID   *int    `json:"owner_id,omitempty"`
	Key       *string `json:"key,omitempty"`
	Value     *string `json:"value,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
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
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
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
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
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
