package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var emailIntegrationSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of this Email integration.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "The name of this Email integration.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the email integration.",
		Type:        schema.TypeString,
		Optional:    true,
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
	"recovery_period": {
		Description: "How long the integration must be up to automatically mark an incident as resolved after being down.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
	"paused": {
		Description: "Set to true to pause monitoring - we won't notify you about downtime. Set to false to resume monitoring.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"email_address": {
		Description: "The email address we expect emails to receive at.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"started_rule_type": {
		Description: "Should an incident be started for all emails, those satisfying all started_rules, or those satisfying any of them. Valid values are unused, all, or any",
		Type:        schema.TypeString,
		Required:    true,
	},
	"acknowledged_rule_type": {
		Description: "Should an incident be acknowledged for all emails, those satisfying all acknowledged_rules, or those satisfying any of them. Valid values are unused, all, or any",
		Type:        schema.TypeString,
		Required:    true,
	},
	"resolved_rule_type": {
		Description: "Should an incident be resolved for all emails, those satisfying all resolved_rules, or those satisfying any of them. Valid values are unused, all, or any",
		Type:        schema.TypeString,
		Required:    true,
	},
	"started_rules": {
		Description: "An array of rules to match to start a new incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationRuleSchema},
		Optional:    true,
	},
	"acknowledged_rules": {
		Description: "An array of rules to match to acknowledge an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationRuleSchema},
		Optional:    true,
	},
	"resolved_rules": {
		Description: "An array of rules to match to resolved an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationRuleSchema},
		Optional:    true,
	},
	"cause_field": {
		Description: "A field describing how to extract an incident cause, used as a short description shared with the team member on-call.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
	"started_alert_id_field": {
		Description: "When starting an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
	"acknowledged_alert_id_field": {
		Description: "When acknowledging an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
	"resolved_alert_id_field": {
		Description: "When resolving an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeSet,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
	"other_started_fields": {
		Description: "An array of additional fields, which will be extracted when starting an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
	"other_acknowledged_fields": {
		Description: "An array of additional fields, which will be extracted when acknowledging an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
	"other_resolved_fields": {
		Description: "An array of additional fields, which will be extracted when resolving an incident.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
	},
}

func newEmailIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: emailIntegrationCreate,
		ReadContext:   emailIntegrationRead,
		UpdateContext: emailIntegrationUpdate,
		DeleteContext: emailIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/email-integrations/",
		Schema:      emailIntegrationSchema,
	}
}

type emailIntegration struct {
	Id                       *int                `json:"id,omitempty"`
	Name                     *string             `json:"name,omitempty"`
	PolicyId                 *string             `json:"policy_id,omitempty"`
	Call                     *bool               `json:"call,omitempty"`
	SMS                      *bool               `json:"sms,omitempty"`
	Email                    *bool               `json:"email,omitempty"`
	Push                     *bool               `json:"push,omitempty"`
	TeamWait                 *int                `json:"team_wait,omitempty"`
	RecoveryPeriod           *int                `json:"recovery_period,omitempty"`
	Paused                   *bool               `json:"paused,omitempty"`
	EmailAddress             *string             `json:"email_address,omitempty"`
	StartedRuleType          *string             `json:"started_rule_type,omitempty"`
	AcknowledgedRuleType     *string             `json:"acknowledged_rule_type,omitempty"`
	ResolvedRuleType         *string             `json:"resolved_rule_type,omitempty"`
	StartedRules             *[]integrationRule  `json:"started_rules,omitempty"`
	AcknowledgedRules        *[]integrationRule  `json:"acknowledged_rules,omitempty"`
	ResolvedRules            *[]integrationRule  `json:"resolved_rules,omitempty"`
	CauseField               *integrationField   `json:"cause_field,omitempty"`
	StartedAlertIdField      *integrationField   `json:"started_alert_id_field,omitempty"`
	AcknowledgedAlertIdField *integrationField   `json:"acknowledged_alert_id_field,omitempty"`
	ResolvedAlertIdField     *integrationField   `json:"resolved_alert_id_field,omitempty"`
	OtherStartedFields       *[]integrationField `json:"other_started_fields,omitempty"`
	OtherAcknowledgedFields  *[]integrationField `json:"other_acknowledged_fields,omitempty"`
	OtherResolvedFields      *[]integrationField `json:"other_resolved_fields,omitempty"`
	TeamName                 *string             `json:"team_name,omitempty"`
}

type emailIntegrationHTTPResponse struct {
	Data struct {
		ID         string           `json:"id"`
		Attributes emailIntegration `json:"attributes"`
	} `json:"data"`
}

func emailIntegrationRef(in *emailIntegration) []struct {
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
		{k: "team_wait", v: &in.TeamWait},
		{k: "recovery_period", v: &in.RecoveryPeriod},
		{k: "paused", v: &in.Paused},
		{k: "email_address", v: &in.EmailAddress},
		{k: "started_rule_type", v: &in.StartedRuleType},
		{k: "acknowledged_rule_type", v: &in.AcknowledgedRuleType},
		{k: "resolved_rule_type", v: &in.ResolvedRuleType},
		{k: "started_rules", v: &in.StartedRules},
		{k: "acknowledged_rules", v: &in.AcknowledgedRules},
		{k: "resolved_rules", v: &in.ResolvedRules},
		{k: "cause_field", v: &in.CauseField},
		{k: "started_alert_id_field", v: &in.StartedAlertIdField},
		{k: "acknowledged_alert_id_field", v: &in.AcknowledgedAlertIdField},
		{k: "resolved_alert_id_field", v: &in.ResolvedAlertIdField},
		{k: "other_started_fields", v: &in.OtherStartedFields},
		{k: "other_acknowledged_fields", v: &in.OtherAcknowledgedFields},
		{k: "other_resolved_fields", v: &in.OtherResolvedFields},
	}
}

func emailIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in emailIntegration

	for _, e := range emailIntegrationRef(&in) {
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
	var out emailIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/email-integrations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return emailIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func emailIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out emailIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/email-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return emailIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func emailIntegrationCopyAttrs(d *schema.ResourceData, in *emailIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range emailIntegrationRef(in) {
		if !isFieldAttribute(e.k) {
			value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
			if err := d.Set(e.k, value); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		}
	}
	return derr
}

func emailIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in emailIntegration
	var out policyHTTPResponse
	for _, e := range emailIntegrationRef(&in) {
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

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/email-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func emailIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/email-integrations/%s", url.PathEscape(d.Id())))
}
