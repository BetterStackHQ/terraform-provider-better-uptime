package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var prometheusIntegrationSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			if d.Id() == "" {
				return false
			}
			return true
		},
	},
	"id": {
		Description: "The ID of the AWS CloudWatch Integration.",
		Type:        schema.TypeString,
		Computed:    true,
		Optional:    true,
	},
	"name": {
		Description: "The name of the AWS CloudWatch Integration.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the AWS CloudWatch integration.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"call": {
		Description: "Do we call the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"sms": {
		Description: "Do we send an SMS to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"email": {
		Description: "Do we send an email to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"push": {
		Description: "Do we send a push notification to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"team_wait": {
		Description: "How long we wait before escalating the incident alert to the team. In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"recovery_period": {
		Description: "How long the alert must be up to automatically mark an incident as resolved. In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
	"paused": {
		Description: "Is the AWS CloudWatch integration paused.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"webhook_url": {
		Description: "The webhook URL for the AWS CloudWatch integration.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

func newPrometheusIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: prometheusIntegrationCreate,
		ReadContext:   prometheusIntegrationRead,
		UpdateContext: prometheusIntegrationUpdate,
		DeleteContext: prometheusIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/prometheus-integrations/",
		Schema:      prometheusIntegrationSchema,
	}
}

type prometheusIntegration struct {
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

type prometheusIntegrationHTTPResponse struct {
	Data struct {
		ID         string                `json:"id"`
		Attributes prometheusIntegration `json:"attributes"`
	} `json:"data"`
}

func prometheusIntegrationRef(in *prometheusIntegration) []struct {
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
		{k: "team_name", v: &in.TeamName},
	}
}

func prometheusIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in prometheusIntegration
	for _, e := range prometheusIntegrationRef(&in) {
		if e.k == "request_headers" {
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
	var out prometheusIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/prometheus-integrations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return prometheusIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func prometheusIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out prometheusIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/prometheus-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return prometheusIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func prometheusIntegrationCopyAttrs(d *schema.ResourceData, in *prometheusIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range prometheusIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func prometheusIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in prometheusIntegration
	var out prometheusIntegrationHTTPResponse
	for _, e := range prometheusIntegrationRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/prometheus-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func prometheusIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/prometheus-integrations/%s", url.PathEscape(d.Id())))
}
