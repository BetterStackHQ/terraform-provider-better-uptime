package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var googleMonitoringIntegrationSchema = map[string]*schema.Schema{
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
		Description: "The ID of the Google Monitoring Integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "The name of the Google Monitoring Integration.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the Google Monitoring integration.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"call": {
		Description: "Do we call the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"sms": {
		Description: "Do we send an SMS to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"email": {
		Description: "Do we send an email to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"push": {
		Description: "Do we send a push notification to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"team_wait": {
		Description: "How long we wait before escalating the incident alert to the team. In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"recovery_period": {
		Description: "How long the alert must be up to automatically mark an incident as resolved. In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"paused": {
		Description: "Is the Google Monitoring integration paused.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"webhook_url": {
		Description: "The webhook URL for the Google Monitoring integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newGoogleMonitoringIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: googleMonitoringIntegrationCreate,
		ReadContext:   googleMonitoringIntegrationRead,
		UpdateContext: googleMonitoringIntegrationUpdate,
		DeleteContext: googleMonitoringIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateRequestHeaders,
		Description:   "https://betterstack.com/docs/uptime/api/google-monitoring-integrations/",
		Schema:        googleMonitoringIntegrationSchema,
	}
}

type googleMonitoringIntegration struct {
	ID             *string `json:"id,omitempty"`
	Name           *string `json:"name,omitempty"`
	PolicyID       *int    `json:"policy_id,omitempty"`
	Call           *bool   `json:"call,omitempty"`
	SMS            *bool   `json:"sms,omitempty"`
	Email          *bool   `json:"email,omitempty"`
	Push           *bool   `json:"push,omitempty"`
	TeamWait       *int    `json:"team_wait,omitempty"`
	RecoveryPeriod *int    `json:"recovery_period,omitempty"`
	Paused         *bool   `json:"paused,omitempty"`
	WebhookURL     *string `json:"webhook_url,omitempty"`
	TeamName       *string `json:"team_name,omitempty"`
}

type googleMonitoringIntegrationHTTPResponse struct {
	Data struct {
		ID         string                      `json:"id"`
		Attributes googleMonitoringIntegration `json:"attributes"`
	} `json:"data"`
}

func googleMonitoringIntegrationRef(in *googleMonitoringIntegration) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "id", v: &in.ID},
		{k: "name", v: &in.Name},
		{k: "policy_id", v: &in.PolicyID},
		{k: "call", v: &in.Call},
		{k: "sms", v: &in.SMS},
		{k: "email", v: &in.Email},
		{k: "push", v: &in.Push},
		{k: "team_wait", v: &in.TeamWait},
		{k: "recovery_period", v: &in.RecoveryPeriod},
		{k: "paused", v: &in.Paused},
		{k: "webhook_url", v: &in.WebhookURL},
	}
}

func googleMonitoringIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in googleMonitoringIntegration
	for _, e := range googleMonitoringIntegrationRef(&in) {
		if e.k == "request_headers" {
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
	var out googleMonitoringIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/google-monitoring-integrations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return googleMonitoringIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func googleMonitoringIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out googleMonitoringIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/google-monitoring-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return googleMonitoringIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func googleMonitoringIntegrationCopyAttrs(d *schema.ResourceData, in *googleMonitoringIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range googleMonitoringIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func googleMonitoringIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in googleMonitoringIntegration
	var out googleMonitoringIntegrationHTTPResponse
	for _, e := range googleMonitoringIntegrationRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/google-monitoring-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func googleMonitoringIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/google-monitoring-integrations/%s", url.PathEscape(d.Id())))
}
