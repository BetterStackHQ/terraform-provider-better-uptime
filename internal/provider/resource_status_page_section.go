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

var statusPageSectionSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Status Page Section.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"status_page_id": {
		Description: "The ID of the Status Page.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"name": {
		Description: "The section name displayed publicly on your status page.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	// TODO: add 'effective_position' computed property?
	"position": {
		Description: "The position of this section on your status page, indexed from zero. If you don't specify a position, we add the section to the end of the status page. When you specify a position of an existing section, we add the section to this position and shift sections below to accommodate.",
		Type:        schema.TypeInt,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return !d.HasChange(k)
		},
	},
}

func newStatusPageSectionResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: statusPageSectionCreate,
		ReadContext:   statusPageSectionRead,
		UpdateContext: statusPageSectionUpdate,
		DeleteContext: statusPageSectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				split := strings.SplitN(d.Id(), "/", 2)
				if len(split) != 2 {
					return nil, errors.New("betteruptime_status_page_section can be imported via \"status_page_id/id\" only (e.g. \"0/1\")")
				}
				if err := d.Set("status_page_id", split[0]); err != nil {
					return nil, err
				}
				d.SetId(split[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: "https://betterstack.com/docs/uptime/api/status-page-sections/",
		Schema:      statusPageSectionSchema,
	}
}

type statusPageSection struct {
	Name          *string `json:"name,omitempty"`
	Position      *int    `json:"position,omitempty"`
	FixedPosition *bool   `json:"fixed_position,omitempty"`
}

type statusPageSectionHTTPResponse struct {
	Data struct {
		ID         string            `json:"id"`
		Attributes statusPageSection `json:"attributes"`
	} `json:"data"`
}

func statusPageSectionRef(in *statusPageSection) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "position", v: &in.Position},
	}
}

func statusPageSectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPageSection
	for _, e := range statusPageSectionRef(&in) {
		load(d, e.k, e.v)
	}
	in.FixedPosition = truePtr()
	statusPageID := d.Get("status_page_id").(string)
	var out statusPageSectionHTTPResponse
	if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/sections", url.PathEscape(statusPageID)), &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return statusPageSectionCopyAttrs(d, &out.Data.Attributes)
}

func statusPageSectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	statusPageID := d.Get("status_page_id").(string)
	var out statusPageSectionHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/sections/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return statusPageSectionCopyAttrs(d, &out.Data.Attributes)
}

func statusPageSectionCopyAttrs(d *schema.ResourceData, in *statusPageSection) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range statusPageSectionRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func statusPageSectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPageSection
	var out policyHTTPResponse
	for _, e := range statusPageSectionRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	in.FixedPosition = truePtr()
	statusPageID := d.Get("status_page_id").(string)
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/sections/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())), &in, &out)
}

func statusPageSectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	statusPageID := d.Get("status_page_id").(string)
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/sections/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())))
}
