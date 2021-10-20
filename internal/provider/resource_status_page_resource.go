package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var statusPageResourceSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Status Page Resource.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"status_page_id": {
		Description: "The ID of the Status Page.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"status_page_section_id": {
		Description: "The ID of the Status Page Section. If you don't specify a status_page_section_id, we add the resource to the first section. If there are no sections in the status page yet, one will be automatically created for you.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"resource_id": {
		Description: "The ID of the resource you are adding.",
		Type:        schema.TypeInt,
		Required:    true,
	},
	"resource_type": {
		Description: "The type of the resource you are adding. Available values: Monitor, Heartbeat, WebhookIntegration, EmailIntegration.",
		Type:        schema.TypeString,
		Required:    true,
		// TODO: ValidateDiagFunc: validation.StringInSlice
	},
	"public_name": {
		Description: "The resource name displayed publicly on your status page.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"explanation": {
		Description: "A detailed text displayed as a help icon.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"history": {
		Description: "Do you want to show the 90-day incident history for this item?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	// TODO: add 'effective_position' computed property?
	"position": {
		Description: "The position of this resource on your status page, indexed from zero. If you don't specify a position, we add the resource to the end of the status page. When you specify a position of an existing resource, we add the resource to this position and shift resources below to accommodate.",
		Type:        schema.TypeInt,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return !d.HasChange(k)
		},
	},
}

func newStatusPageResourceResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: statusPageResourceCreate,
		ReadContext:   statusPageResourceRead,
		UpdateContext: statusPageResourceUpdate,
		DeleteContext: statusPageResourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				split := strings.SplitN(d.Id(), "/", 2)
				if len(split) != 2 {
					return nil, errors.New("betteruptime_status_page_resource can be imported via \"status_page_id/id\" only (e.g. \"0/1\")")
				}
				if err := d.Set("status_page_id", split[0]); err != nil {
					return nil, err
				}
				d.SetId(split[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: "https://docs.betteruptime.com/api/status-page-resources-api",
		Schema:      statusPageResourceSchema,
	}
}

type statusPageResource struct {
	StatusPageSectionID *int    `json:"status_page_section_id,omitempty"`
	ResourceID          *int    `json:"resource_id,omitempty"`
	ResourceType        *string `json:"resource_type,omitempty"`
	PublicName          *string `json:"public_name,omitempty"`
	Explanation         *string `json:"explanation,omitempty"`
	History             *bool   `json:"history,omitempty"`
	Position            *int    `json:"position,omitempty"`
}

type statusPageResourceHTTPResponse struct {
	Data struct {
		ID         string             `json:"id"`
		Attributes statusPageResource `json:"attributes"`
	} `json:"data"`
}

func statusPageResourceRef(in *statusPageResource) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "status_page_section_id", v: &in.StatusPageSectionID},
		{k: "resource_id", v: &in.ResourceID},
		{k: "resource_type", v: &in.ResourceType},
		{k: "public_name", v: &in.PublicName},
		{k: "explanation", v: &in.Explanation},
		{k: "history", v: &in.History},
		{k: "position", v: &in.Position},
	}
}

func statusPageResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPageResource
	for _, e := range statusPageResourceRef(&in) {
		load(d, e.k, e.v)
	}
	statusPageID := d.Get("status_page_id").(string)
	var out statusPageResourceHTTPResponse
	if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources?fixed_position=true", url.PathEscape(statusPageID)), &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return statusPageResourceCopyAttrs(d, &out.Data.Attributes)
}

func statusPageResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	statusPageID := d.Get("status_page_id").(string)
	var out statusPageResourceHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return statusPageResourceCopyAttrs(d, &out.Data.Attributes)
}

func statusPageResourceCopyAttrs(d *schema.ResourceData, in *statusPageResource) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range statusPageResourceRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func statusPageResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPageResource
	for _, e := range statusPageResourceRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	statusPageID := d.Get("status_page_id").(string)
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources/%s?fixed_position=true", url.PathEscape(statusPageID), url.PathEscape(d.Id())), &in)
}

func statusPageResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	statusPageID := d.Get("status_page_id").(string)
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())))
}
