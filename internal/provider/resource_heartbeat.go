package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var heartbeatSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this heartbeat.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "A name of the service for this heartbeat.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"url": {
		Description: "The url of this heartbeat.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"period": {
		Description: "How often should we expect this heartbeat? In seconds. Minimum value: 30 seconds",
		Type:        schema.TypeInt,
		Required:    true,
		// TODO: ValidateDiagFunc
	},
	"grace": {
		Description: "Heartbeats can fluctuate; specify this value to control what is still acceptable. Minimum value: 0 seconds. We recommend setting this to approx. 20% of period",
		Type:        schema.TypeInt,
		Required:    true,
		// TODO: ValidateDiagFunc
	},
	"call": {
		Description: "Should we call the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"sms": {
		Description: "Should we send an SMS to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"email": {
		Description: "Should we send an email to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"push": {
		Description: "Should we send a push notification to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"team_wait": {
		Description: "How long to wait before escalating the incident alert to the team. Leave blank to disable escalating to the entire team.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"heartbeat_group_id": {
		Description: "Set this attribute if you want to add this heartbeat to a heartbeat group..",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"sort_index": {
		Description: "An index controlling the position of a heartbeat in the heartbeat group.",
		Type:        schema.TypeInt,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return !d.HasChange(k)
		},
	},
	"paused": {
		Description: "Set to true to pause monitoring — we won't notify you about downtime. Set to false to resume monitoring.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"paused_at": {
		Description: "The time when this heartbeat was paused.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"policy_id": {
		Description: "Set the escalation policy for the heartbeat.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"status": {
		Description: "The status of this heartbeat.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"created_at": {
		Description: "The time when this heartbeat was created.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this heartbeat was updated.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

func newHeartbeatResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: heartbeatCreate,
		ReadContext:   heartbeatRead,
		UpdateContext: heartbeatUpdate,
		DeleteContext: heartbeatDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/hearbeats/",
		Schema:      heartbeatSchema,
	}
}

type heartbeat struct {
	Name             *string `json:"name,omitempty"`
	Url              *string `json:"url,omitempty"`
	Period           *int    `json:"period,omitempty"`
	Grace            *int    `json:"grace,omitempty"`
	Call             *bool   `json:"call,omitempty"`
	SMS              *bool   `json:"sms,omitempty"`
	Email            *bool   `json:"email,omitempty"`
	Push             *bool   `json:"push,omitempty"`
	TeamWait         *int    `json:"team_wait,omitempty"`
	HeartbeatGroupID *int    `json:"heartbeat_group_id,omitempty"`
	SortIndex        *int    `json:"sort_index,omitempty"`
	Paused           *bool   `json:"paused,omitempty"`
	PausedAt         *string `json:"paused_at,omitempty"`
	PolicyID         *string `json:"policy_id"`
	Status           *string `json:"status,omitempty"`
	CreatedAt        *string `json:"created_at,omitempty"`
	UpdatedAt        *string `json:"updated_at,omitempty"`
}

type heartbeatHTTPResponse struct {
	Data struct {
		ID         string    `json:"id"`
		Attributes heartbeat `json:"attributes"`
	} `json:"data"`
}

func heartbeatRef(in *heartbeat) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "url", v: &in.Url},
		{k: "period", v: &in.Period},
		{k: "grace", v: &in.Grace},
		{k: "call", v: &in.Call},
		{k: "sms", v: &in.SMS},
		{k: "email", v: &in.Email},
		{k: "push", v: &in.Push},
		{k: "team_wait", v: &in.TeamWait},
		{k: "heartbeat_group_id", v: &in.HeartbeatGroupID},
		{k: "sort_index", v: &in.SortIndex},
		{k: "paused", v: &in.Paused},
		{k: "paused_at", v: &in.PausedAt},
		{k: "policy_id", v: &in.PolicyID},
		{k: "status", v: &in.Status},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func heartbeatCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in heartbeat
	for _, e := range heartbeatRef(&in) {
		load(d, e.k, e.v)
	}
	var out heartbeatHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/heartbeats", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return heartbeatCopyAttrs(d, &out.Data.Attributes)
}

func heartbeatRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out heartbeatHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/heartbeats/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return heartbeatCopyAttrs(d, &out.Data.Attributes)
}

func heartbeatCopyAttrs(d *schema.ResourceData, in *heartbeat) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range heartbeatRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func heartbeatUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in heartbeat
	var out policyHTTPResponse
	for _, e := range heartbeatRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/heartbeats/%s", url.PathEscape(d.Id())), &in, &out)
}

func heartbeatDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/heartbeats/%s", url.PathEscape(d.Id())))
}
