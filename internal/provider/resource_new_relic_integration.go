package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var newRelicIntegrationSchema = map[string]*schema.Schema{
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
		Description: "The ID of the AWS CloudWatch Integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "The name of the AWS CloudWatch Integration.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the AWS CloudWatch integration.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"call": {
		Description: "Whether to call when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"sms": {
		Description: "Whether to send an SMS when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"email": {
		Description: "Whether to send an email when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"push": {
		Description: "Whether to send a push notification when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"critical_alert": {
		Description: "Whether to send a critical push notification that ignores the mute switch and Do not Disturb mode when a new incident is created.",
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
	"alerting_rule": {
		Description: "Should we alert only on alarms, or on both alarms and warnings. Possible values: alert, alert_and_warn.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"paused": {
		Description: "Is the AWS CloudWatch integration paused.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"webhook_url": {
		Description: "The webhook URL for the AWS CloudWatch integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newNewRelicIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: newRelicIntegrationCreate,
		ReadContext:   newRelicIntegrationRead,
		UpdateContext: newRelicIntegrationUpdate,
		DeleteContext: newRelicIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateRequestHeaders,
		Description:   "https://betterstack.com/docs/uptime/api/new-relic-integrations/",
		Schema:        newRelicIntegrationSchema,
	}
}

type newRelicIntegration struct {
	ID             *string `json:"id,omitempty"`
	Name           *string `json:"name,omitempty"`
	PolicyID       *int    `json:"policy_id,omitempty"`
	Call           *bool   `json:"call,omitempty"`
	SMS            *bool   `json:"sms,omitempty"`
	Email          *bool   `json:"email,omitempty"`
	Push           *bool   `json:"push,omitempty"`
	CriticalAlert  *bool   `json:"critical_alert,omitempty"`
	TeamWait       *int    `json:"team_wait,omitempty"`
	RecoveryPeriod *int    `json:"recovery_period,omitempty"`
	AlertingRule   *string `json:"alerting_rule,omitempty"`
	Paused         *bool   `json:"paused,omitempty"`
	WebhookURL     *string `json:"webhook_url,omitempty"`
	TeamName       *string `json:"team_name,omitempty"`
}

type newRelicIntegrationHTTPResponse struct {
	Data struct {
		ID         string              `json:"id"`
		Attributes newRelicIntegration `json:"attributes"`
	} `json:"data"`
}

func newRelicIntegrationRef(in *newRelicIntegration) []struct {
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
		{k: "critical_alert", v: &in.CriticalAlert},
		{k: "team_wait", v: &in.TeamWait},
		{k: "recovery_period", v: &in.RecoveryPeriod},
		{k: "alerting_rule", v: &in.AlertingRule},
		{k: "paused", v: &in.Paused},
		{k: "webhook_url", v: &in.WebhookURL},
	}
}

func newRelicIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in newRelicIntegration
	for _, e := range newRelicIntegrationRef(&in) {
		if e.k == "request_headers" {
			if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
				return diag.FromErr(err)
			}
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
	var out newRelicIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/new-relic-integrations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return newRelicIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func newRelicIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out newRelicIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/new-relic-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return newRelicIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func newRelicIntegrationCopyAttrs(d *schema.ResourceData, in *newRelicIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range newRelicIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func newRelicIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in newRelicIntegration
	var out newRelicIntegrationHTTPResponse
	for _, e := range newRelicIntegrationRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
					return diag.FromErr(err)
				}
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/new-relic-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func newRelicIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/new-relic-integrations/%s", url.PathEscape(d.Id())))
}
