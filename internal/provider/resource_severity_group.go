package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var severityGroupSchema = map[string]*schema.Schema{
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
		Description: "The ID of this severity group.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "A name of the group that you can see in the dashboard.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"sort_index": {
		Description: "Set sort_index to specify how to sort your severity groups.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return !d.HasChange(k)
		},
	},
	"created_at": {
		Description: "The time when this severity group was created.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this severity group was updated.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newSeverityGroupResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: severityGroupCreate,
		ReadContext:   severityGroupRead,
		UpdateContext: severityGroupUpdate,
		DeleteContext: severityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/urgency-groups/",
		Schema:      severityGroupSchema,
	}
}

type severityGroup struct {
	Name      *string `json:"name,omitempty"`
	SortIndex *int    `json:"sort_index,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	TeamName  *string `json:"team_name,omitempty"`
}

type severityGroupHTTPResponse struct {
	Data struct {
		ID         string        `json:"id"`
		Attributes severityGroup `json:"attributes"`
	} `json:"data"`
}

func severityGroupRef(in *severityGroup) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "sort_index", v: &in.SortIndex},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func severityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in severityGroup
	for _, e := range severityGroupRef(&in) {
		load(d, e.k, e.v)
	}
	load(d, "team_name", &in.TeamName)
	var out severityGroupHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/urgency-groups", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return severityGroupCopyAttrs(d, &out.Data.Attributes)
}

func severityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out severityGroupHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/urgency-groups/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return severityGroupCopyAttrs(d, &out.Data.Attributes)
}

func severityGroupCopyAttrs(d *schema.ResourceData, in *severityGroup) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range severityGroupRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func severityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in severityGroup
	var out severityHTTPResponse
	for _, e := range severityGroupRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/urgency-groups/%s", url.PathEscape(d.Id())), &in, &out)
}

func severityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/urgency-groups/%s", url.PathEscape(d.Id())))
}
