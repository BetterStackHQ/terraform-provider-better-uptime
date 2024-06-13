package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var severitySchema = map[string]*schema.Schema{
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
		Description: "The ID of this Severity.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "The name of this Severity.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"sms": {
		Description: "Whether to send SMS when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"call": {
		Description: "Whether to call when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"email": {
		Description: "Whether to send email when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"push": {
		Description: "Whether to send push notification when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
}

func newSeverityResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: severityCreate,
		ReadContext:   severityRead,
		UpdateContext: severityUpdate,
		DeleteContext: severityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/list-all-severities/",
		Schema:      severitySchema,
	}
}

type severity struct {
	Id       *int    `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	SMS      *bool   `json:"sms,omitempty"`
	Call     *bool   `json:"call,omitempty"`
	Email    *bool   `json:"email,omitempty"`
	Push     *bool   `json:"push,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
}

type severityHTTPResponse struct {
	Data struct {
		ID         string   `json:"id"`
		Attributes severity `json:"attributes"`
	} `json:"data"`
}

func severityRef(in *severity) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "sms", v: &in.SMS},
		{k: "call", v: &in.Call},
		{k: "email", v: &in.Email},
		{k: "push", v: &in.Push},
	}
}

func severityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in severity
	for _, e := range severityRef(&in) {
		load(d, e.k, e.v)
	}
	load(d, "team_name", &in.TeamName)
	var out severityHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/urgencies", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return severityCopyAttrs(d, &out.Data.Attributes)
}

func severityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out severityHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/urgencies/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return severityCopyAttrs(d, &out.Data.Attributes)
}

func severityCopyAttrs(d *schema.ResourceData, in *severity) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range severityRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func severityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in severity
	var out severityHTTPResponse
	for _, e := range severityRef(&in) {
		load(d, e.k, e.v)
	}

	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/urgencies/%s", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return severityCopyAttrs(d, &out.Data.Attributes)
}

func severityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/urgencies/%s", url.PathEscape(d.Id())))
}
