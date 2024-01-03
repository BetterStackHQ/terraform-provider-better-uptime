package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var awsCloudWatchIntegrationSchema = map[string]*schema.Schema{
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

func newAwsCloudWatchIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: awsCloudWatchIntegrationCreate,
		ReadContext:   awsCloudWatchIntegrationRead,
		UpdateContext: awsCloudWatchIntegrationUpdate,
		DeleteContext: awsCloudWatchIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/aws-cloudwatch-integrations/",
		Schema:      awsCloudWatchIntegrationSchema,
	}
}

type awsCloudWatchIntegration struct {
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
}

type awsCloudWatchIntegrationHTTPResponse struct {
	Data struct {
		ID         string                   `json:"id"`
		Attributes awsCloudWatchIntegration `json:"attributes"`
	} `json:"data"`
}

func awsCloudWatchIntegrationRef(in *awsCloudWatchIntegration) []struct {
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

func awsCloudWatchIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in awsCloudWatchIntegration
	for _, e := range awsCloudWatchIntegrationRef(&in) {
		if e.k == "request_headers" {
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
	var out awsCloudWatchIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/aws-cloudwatch-integrations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return awsCloudWatchIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func awsCloudWatchIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out awsCloudWatchIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/aws-cloudwatch-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return awsCloudWatchIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func awsCloudWatchIntegrationCopyAttrs(d *schema.ResourceData, in *awsCloudWatchIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range awsCloudWatchIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func awsCloudWatchIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in awsCloudWatchIntegration
	var out awsCloudWatchIntegrationHTTPResponse
	for _, e := range awsCloudWatchIntegrationRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/aws-cloudwatch-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func awsCloudWatchIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/aws-cloudwatch-integrations/%s", url.PathEscape(d.Id())))
}
