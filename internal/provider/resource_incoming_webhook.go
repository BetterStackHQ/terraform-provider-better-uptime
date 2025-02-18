package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var incomingWebhookSchema = map[string]*schema.Schema{
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
		Description: "The ID of this incoming webhook.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "The name of this incoming webhook.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the incoming webhook.",
		Type:        schema.TypeString,
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
		Description: "How long to wait before escalating the incident alert to the team. Leave blank to disable escalating to the entire team.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"recovery_period": {
		Description: "How long the integration must be up to automatically mark an incident as resolved after being down.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"paused": {
		Description: "Set to true to pause monitoring - we won't notify you about downtime. Set to false to resume monitoring.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"url": {
		Description: "The url at which we expect to receive the webhook.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"sample_query_string": {
		Description: "Sample query string of the webhook (without the leading ?). Used only to make the configuration easier.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"sample_headers": {
		Description: "Sample request HTTP headers the webhook (separated by a newline). Used only to make the configuration easier.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"sample_body": {
		Description: "Sample request body the webhook. Used only to make the configuration easier.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"started_rule_type": {
		Description: "Should an incident be started for all webhooks, those satisfying all started_rules, or those satisfying any of them. Valid values are unused, all, or any",
		Type:        schema.TypeString,
		Required:    true,
	},
	"acknowledged_rule_type": {
		Description: "Should an incident be acknowledged for all webhooks, those satisfying all acknowledged_rules, or those satisfying any of them. Valid values are unused, all, or any",
		Type:        schema.TypeString,
		Required:    true,
	},
	"resolved_rule_type": {
		Description: "Should an incident be resolved for all webhooks, those satisfying all resolved_rules, or those satisfying any of them. Valid values are unused, all, or any",
		Type:        schema.TypeString,
		Required:    true,
	},
	"started_rules": {
		Description: "An array of rules to match to start a new incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationRuleSchema},
		Optional:    true,
		Computed:    true,
	},
	"acknowledged_rules": {
		Description: "An array of rules to match to acknowledge an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationRuleSchema},
		Optional:    true,
		Computed:    true,
	},
	"resolved_rules": {
		Description: "An array of rules to match to resolved an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationRuleSchema},
		Optional:    true,
		Computed:    true,
	},
	"cause_field": {
		Description: "A field describing how to extract an incident cause, used as a short description shared with the team member on-call.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"title_field": {
		Description: "An optional field describing how to extract a customized incident title.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    false,
	},
	"started_alert_id_field": {
		Description: "When starting an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"acknowledged_alert_id_field": {
		Description: "When acknowledging an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"resolved_alert_id_field": {
		Description: "When resolving an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"other_started_fields": {
		Description: "An array of additional fields, which will be extracted when starting an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"other_acknowledged_fields": {
		Description: "An array of additional fields, which will be extracted when acknowledging an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"other_resolved_fields": {
		Description: "An array of additional fields, which will be extracted when resolving an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
	},
	"created_at": {
		Description: "The time when this incoming webhook was created.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this incoming webhook was updated.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newIncomingWebhookResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: incomingWebhookCreate,
		ReadContext:   incomingWebhookRead,
		UpdateContext: incomingWebhookUpdate,
		DeleteContext: incomingWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/list-all-incoming-webhooks/",
		Schema:      incomingWebhookSchema,
	}
}

type incomingWebhook struct {
	Id                       *int                `json:"id,omitempty"`
	Name                     *string             `json:"name,omitempty"`
	PolicyId                 *string             `json:"policy_id,omitempty"`
	Call                     *bool               `json:"call,omitempty"`
	SMS                      *bool               `json:"sms,omitempty"`
	Email                    *bool               `json:"email,omitempty"`
	Push                     *bool               `json:"push,omitempty"`
	CriticalAlert            *bool               `json:"critical_alert,omitempty"`
	TeamWait                 *int                `json:"team_wait,omitempty"`
	RecoveryPeriod           *int                `json:"recovery_period,omitempty"`
	Paused                   *bool               `json:"paused,omitempty"`
	Url                      *string             `json:"url,omitempty"`
	SampleQueryString        *string             `json:"sample_query_string,omitempty"`
	SampleHeaders            *string             `json:"sample_headers,omitempty"`
	SampleBody               *string             `json:"sample_body,omitempty"`
	StartedRuleType          *string             `json:"started_rule_type,omitempty"`
	AcknowledgedRuleType     *string             `json:"acknowledged_rule_type,omitempty"`
	ResolvedRuleType         *string             `json:"resolved_rule_type,omitempty"`
	StartedRules             *[]integrationRule  `json:"started_rules,omitempty"`
	AcknowledgedRules        *[]integrationRule  `json:"acknowledged_rules,omitempty"`
	ResolvedRules            *[]integrationRule  `json:"resolved_rules,omitempty"`
	CauseField               *integrationField   `json:"cause_field,omitempty"`
	TitleField               *integrationField   `json:"title_field"`
	StartedAlertIdField      *integrationField   `json:"started_alert_id_field,omitempty"`
	AcknowledgedAlertIdField *integrationField   `json:"acknowledged_alert_id_field,omitempty"`
	ResolvedAlertIdField     *integrationField   `json:"resolved_alert_id_field,omitempty"`
	OtherStartedFields       *[]integrationField `json:"other_started_fields,omitempty"`
	OtherAcknowledgedFields  *[]integrationField `json:"other_acknowledged_fields,omitempty"`
	OtherResolvedFields      *[]integrationField `json:"other_resolved_fields,omitempty"`
	TeamName                 *string             `json:"team_name,omitempty"`
	CreatedAt                *string             `json:"created_at,omitempty"`
	UpdatedAt                *string             `json:"updated_at,omitempty"`
}

type incomingWebhookHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes incomingWebhook `json:"attributes"`
	} `json:"data"`
}

func incomingWebhookRef(in *incomingWebhook) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "policy_id", v: &in.PolicyId},
		{k: "call", v: &in.Call},
		{k: "sms", v: &in.SMS},
		{k: "email", v: &in.Email},
		{k: "push", v: &in.Push},
		{k: "critical_alert", v: &in.CriticalAlert},
		{k: "team_wait", v: &in.TeamWait},
		{k: "recovery_period", v: &in.RecoveryPeriod},
		{k: "paused", v: &in.Paused},
		{k: "url", v: &in.Url},
		{k: "sample_query_string", v: &in.SampleQueryString},
		{k: "sample_headers", v: &in.SampleHeaders},
		{k: "sample_body", v: &in.SampleBody},
		{k: "started_rule_type", v: &in.StartedRuleType},
		{k: "acknowledged_rule_type", v: &in.AcknowledgedRuleType},
		{k: "resolved_rule_type", v: &in.ResolvedRuleType},
		{k: "started_rules", v: &in.StartedRules},
		{k: "acknowledged_rules", v: &in.AcknowledgedRules},
		{k: "resolved_rules", v: &in.ResolvedRules},
		{k: "cause_field", v: &in.CauseField},
		{k: "title_field", v: &in.TitleField},
		{k: "started_alert_id_field", v: &in.StartedAlertIdField},
		{k: "acknowledged_alert_id_field", v: &in.AcknowledgedAlertIdField},
		{k: "resolved_alert_id_field", v: &in.ResolvedAlertIdField},
		{k: "other_started_fields", v: &in.OtherStartedFields},
		{k: "other_acknowledged_fields", v: &in.OtherAcknowledgedFields},
		{k: "other_resolved_fields", v: &in.OtherResolvedFields},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func incomingWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in incomingWebhook

	for _, e := range incomingWebhookRef(&in) {
		if isRulesAttribute(e.k) {
			loadIntegrationRules(d, e.k, e.v.(**[]integrationRule))
		} else if isFieldAttribute(e.k) {
			loadIntegrationField(d, e.k, e.v.(**integrationField))
		} else if isFieldsAttribute(e.k) {
			loadIntegrationFields(d, e.k, e.v.(**[]integrationField))
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
	var out incomingWebhookHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/incoming-webhooks", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return incomingWebhookCopyAttrs(d, &out.Data.Attributes)
}

func incomingWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out incomingWebhookHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/incoming-webhooks/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return incomingWebhookCopyAttrs(d, &out.Data.Attributes)
}

func incomingWebhookCopyAttrs(d *schema.ResourceData, in *incomingWebhook) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range incomingWebhookRef(in) {
		if !isFieldAttribute(e.k) {
			value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
			if err := d.Set(e.k, value); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		}
	}
	return derr
}

func incomingWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in incomingWebhook
	var out policyHTTPResponse
	for _, e := range incomingWebhookRef(&in) {
		if d.HasChange(e.k) {
			if isRulesAttribute(e.k) {
				loadIntegrationRules(d, e.k, e.v.(**[]integrationRule))
			} else if isFieldAttribute(e.k) {
				loadIntegrationField(d, e.k, e.v.(**integrationField))
			} else if isFieldsAttribute(e.k) {
				loadIntegrationFields(d, e.k, e.v.(**[]integrationField))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/incoming-webhooks/%s", url.PathEscape(d.Id())), &in, &out)
}

func incomingWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/incoming-webhooks/%s", url.PathEscape(d.Id())))
}
