package provider

// Amazon SNS integrations are the same resource shape as incoming webhooks: they share their
// extraction model and their receive URL, which is why this mirrors resource_incoming_webhook.go.
// What is its own is the subscription handshake, and both of its fields are read-only: Better
// Stack writes them when AWS confirms the subscription, so Terraform must not try to manage them.

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var amazonSnsIntegrationSchema = map[string]*schema.Schema{
	"team_name": teamNameSchema(),
	"topic_arn": {
		Description: "The ARN of the Amazon SNS topic this integration is subscribed to. Set by Better Stack when AWS confirms the subscription, so it is read-only.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"subscription_state": {
		Description: "Whether the Amazon SNS subscription has been confirmed: active once AWS has confirmed it, awaiting until then. Set by Better Stack when AWS confirms it, so it is read-only.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"id": {
		Description: "The ID of this Amazon SNS integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "The name of this Amazon SNS integration.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"policy_id": {
		Description: "ID of the escalation policy associated with the Amazon SNS integration.",
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
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
	},
	// Computed, unlike the incoming webhook resource's copy: the API gives a fresh Amazon SNS
	// integration a title extracted from the envelope's Subject. Leaving this Optional-only would
	// send title_field: null on every create, which the API reads as "destroy it" and which would
	// silently drop that default; and it would then show the server's default as a permanent diff.
	"title_field": {
		Description: "An optional field describing how to extract a customized incident title. Defaults to the Amazon SNS envelope's Subject. Omit the block to keep whatever is configured; it cannot be removed through Terraform.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
	},
	"started_alert_id_field": {
		Description: "When starting an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
	},
	"acknowledged_alert_id_field": {
		Description: "When acknowledging an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
	},
	"resolved_alert_id_field": {
		Description: "When resolving an incident, how to extract an alert id, a unique alert identifier which will be used to acknowledge and resolve incidents.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: integrationFieldSchema},
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
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
		Description: "The time when this Amazon SNS integration was created.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this Amazon SNS integration was updated.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
}

func newAmazonSnsIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: amazonSnsIntegrationCreate,
		ReadContext:   amazonSnsIntegrationRead,
		UpdateContext: amazonSnsIntegrationUpdate,
		DeleteContext: amazonSnsIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/list-all-amazon-sns-integrations/",
		CustomizeDiff: customdiff.Sequence(
			validateTeamNameNotChanged,
			validateIntegrationRuleConditions,
			validateSnsEnvelopeTargetField,
		),
		Schema: amazonSnsIntegrationSchema,
	}
}

type amazonSnsIntegration struct {
	Id                       *int                `json:"id,omitempty"`
	TopicArn                 *string             `json:"topic_arn,omitempty"`
	SubscriptionState        *string             `json:"subscription_state,omitempty"`
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
	TitleField               *integrationField   `json:"title_field,omitempty"`
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

type amazonSnsIntegrationHTTPResponse struct {
	Data struct {
		ID         string               `json:"id"`
		Attributes amazonSnsIntegration `json:"attributes"`
	} `json:"data"`
}

func amazonSnsIntegrationRef(in *amazonSnsIntegration) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "topic_arn", v: &in.TopicArn},
		{k: "subscription_state", v: &in.SubscriptionState},
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

func amazonSnsIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in amazonSnsIntegration

	for _, e := range amazonSnsIntegrationRef(&in) {
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
	var out amazonSnsIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/amazon-sns", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return amazonSnsIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func amazonSnsIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out amazonSnsIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/amazon-sns/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return amazonSnsIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func amazonSnsIntegrationCopyAttrs(d *schema.ResourceData, in *amazonSnsIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range amazonSnsIntegrationRef(in) {
		value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
		// Handle field attributes that need special formatting (null or set of 1 element)
		if isFieldAttribute(e.k) {
			if reflect.ValueOf(value).IsNil() {
				value = nil
			} else {
				value = []interface{}{value}
			}
		}
		if err := d.Set(e.k, value); err != nil {
			derr = append(derr, diag.FromErr(err)...)
		}
	}
	return derr
}

func amazonSnsIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in amazonSnsIntegration
	var out amazonSnsIntegrationHTTPResponse
	for _, e := range amazonSnsIntegrationRef(&in) {
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

	// Copying the response back matters more here than for a plain incoming webhook: topic_arn and
	// subscription_state flip on their own when AWS answers the handshake, so an update that did not
	// refresh them would leave an integration confirmed between applies looking unconfirmed in state.
	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/amazon-sns/%s", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	return amazonSnsIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func amazonSnsIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/amazon-sns/%s", url.PathEscape(d.Id())))
}
