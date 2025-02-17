package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var azureIntegrationSchema = map[string]*schema.Schema{
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
		Description: "The ID of the Azure Integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "The name of the Azure Integration.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the Azure integration.",
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
		Description: "Whether to send SMS when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"email": {
		Description: "Whether to send email when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"push": {
		Description: "Whether to send push notification when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"critical_alert": {
		Description: "Whether to send critical alert when a new incident is created.",
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
		Description: "Is the Azure integration paused.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"webhook_url": {
		Description: "The webhook URL for the Azure integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newAzureIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: azureIntegrationCreate,
		ReadContext:   azureIntegrationRead,
		UpdateContext: azureIntegrationUpdate,
		DeleteContext: azureIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateRequestHeaders,
		Description:   "https://betterstack.com/docs/uptime/api/azure-integrations/",
		Schema:        azureIntegrationSchema,
	}
}

type azureIntegration struct {
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
	Paused         *bool   `json:"paused,omitempty"`
	WebhookURL     *string `json:"webhook_url,omitempty"`
	TeamName       *string `json:"team_name,omitempty"`
}

type azureIntegrationHTTPResponse struct {
	Data struct {
		ID         string           `json:"id"`
		Attributes azureIntegration `json:"attributes"`
	} `json:"data"`
}

func azureIntegrationRef(in *azureIntegration) []struct {
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
		{k: "paused", v: &in.Paused},
		{k: "webhook_url", v: &in.WebhookURL},
	}
}

func azureIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in azureIntegration
	for _, e := range azureIntegrationRef(&in) {
		if e.k == "request_headers" {
			if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
				return diag.FromErr(err)
			}
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
	var out azureIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/azure-integrations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return azureIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func azureIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out azureIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/azure-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return azureIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func azureIntegrationCopyAttrs(d *schema.ResourceData, in *azureIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range azureIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func azureIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in azureIntegration
	var out azureIntegrationHTTPResponse
	for _, e := range azureIntegrationRef(&in) {
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

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/azure-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func azureIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/azure-integrations/%s", url.PathEscape(d.Id())))
}
