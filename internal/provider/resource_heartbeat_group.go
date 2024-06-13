package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var heartbeatGroupSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of this Monitor.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"paused": {
		Description: "Set to true to pause monitoring for any existing heartbeats in the group - we won't notify you about downtime. Set to false to resume monitoring for any existing heartbeats in the group.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"name": {
		Description: "A name of the group that you can see in the dashboard.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"sort_index": {
		Description: "Set sort_index to specify how to sort your heartbeat groups.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return !d.HasChange(k)
		},
	},
	"created_at": {
		Description: "The time when this heartbeat group was created.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this heartbeat group was updated.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newHeartbeatGroupResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: heartbeatGroupCreate,
		ReadContext:   heartbeatGroupRead,
		UpdateContext: heartbeatGroupUpdate,
		DeleteContext: heartbeatGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/heartbeat-groups/",
		Schema:      heartbeatGroupSchema,
	}
}

type heartbeatGroup struct {
	Paused    *bool   `json:"paused,omitempty"`
	Name      *string `json:"name,omitempty"`
	SortIndex *int    `json:"sort_index,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	TeamName  *string `json:"team_name,omitempty"`
}

type heartbeatGroupHTTPResponse struct {
	Data struct {
		ID         string         `json:"id"`
		Attributes heartbeatGroup `json:"attributes"`
	} `json:"data"`
}

func heartbeatGroupRef(in *heartbeatGroup) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "paused", v: &in.Paused},
		{k: "name", v: &in.Name},
		{k: "sort_index", v: &in.SortIndex},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func heartbeatGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in heartbeatGroup
	for _, e := range heartbeatGroupRef(&in) {
		load(d, e.k, e.v)
	}
	load(d, "team_name", &in.TeamName)
	var out heartbeatGroupHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/heartbeat-groups", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return heartbeatGroupCopyAttrs(d, &out.Data.Attributes)
}

func heartbeatGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out heartbeatGroupHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/heartbeat-groups/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return heartbeatGroupCopyAttrs(d, &out.Data.Attributes)
}

func heartbeatGroupCopyAttrs(d *schema.ResourceData, in *heartbeatGroup) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range heartbeatGroupRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func heartbeatGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in heartbeatGroup
	var out policyHTTPResponse
	for _, e := range heartbeatGroupRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/heartbeat-groups/%s", url.PathEscape(d.Id())), &in, &out)
}

func heartbeatGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/heartbeat-groups/%s", url.PathEscape(d.Id())))
}
