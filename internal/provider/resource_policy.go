package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
)

var policyStepMemberSchema = map[string]*schema.Schema{
	"type": {
		Description: "Type type of the member to notify during an incident. Can be one of current_on_call, entire_team, all_slack_integrations, all_microsoft_teams_integrations, all_zapier_integrations, all_webhook_integrations, all_splunk_on_call_integrations, user, webhook, slack_integration, microsoft_teams_integration, zapier_webhook.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"id": {
		Description: "The ID of the resource to notify during an incident. Required for user, webhook, slack_integration, microsoft_teams_integration and zapier_webhook member types. This is e.g. the ID of the user to notify when member type is user.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"team_id": {
		Description: "The ID of the team to notify when member team is entire_team. When left empty, the default team for the incident is used.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
}

var policyStepSchema = map[string]*schema.Schema{
	"type": {
		Description: "The type of the step. Can be either escalation or time_branching.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"wait_before": {
		Description: "How long to wait before executing this step since previous step.",
		Type:        schema.TypeInt,
		Required:    true,
	},
	"urgency_id": {
		Description: "Which urgency to use for this step.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"timezone": {
		Description: "What timezone to use when evaluating time based branching rules. Used when step type is branching. The accepted values can be found in the Rails TimeZone documentation. https://api.rubyonrails.org/classes/ActiveSupport/TimeZone.html",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"days": {
		Description: "An array of days during which the branching rule will be executed. Valid values are [\"mon\", \"tue\", \"wed\", \"thu\", \"fri\", \"sat\", \"sun\"]. Used when step type is branching.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"time_from": {
		Description: "A time from which the branching rule will be executed. Use HH:MM format. Used when step type is branching.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"time_to": {
		Description: "A time at which the branching rule will step being executed. Use HH:MM format. Used when step type is branching.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"policy_id": {
		Description: "A policy to executed if the branching rule matches the time of an incident. Used when step type is branching.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"step_members": {
		Description: "An array of escalation policy steps members.",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: policyStepMemberSchema},
		Optional:    true,
	},
}

var policySchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Policy.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "The name of this Policy.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"repeat_count": {
		Description: "How many times should the entire policy be repeated if no one acknowledges the incident.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
	"repeat_delay": {
		Description: "How long in seconds to wait before each repetition.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
	"incident_token": {
		Description: "Incident token that can be used for manually reporting incidents.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"steps": {
		Description: "An array of escalation policy steps",
		Type:        schema.TypeList,
		Elem:        &schema.Resource{Schema: policyStepSchema},
		Required:    true,
	},
}

func newPolicyResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: policyCreate,
		ReadContext:   policyRead,
		UpdateContext: policyUpdate,
		DeleteContext: policyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/list-all-escalation-policies/",
		Schema:      policySchema,
	}
}

type policyStepMember struct {
	Type   *string `mapstructure:"type,omitempty" json:"type,omitempty"`
	Id     *int    `mapstructure:"id,omitempty" json:"id,omitempty"`
	TeamId *int    `mapstructure:"team_id,omitempty" json:"team_id,omitempty"`
}

type policyStep struct {
	Type       *string             `mapstructure:"type,omitempty" json:"type,omitempty"`
	WaitBefore *int                `mapstructure:"wait_before,omitempty" json:"wait_before,omitempty"`
	UrgencyId  *int                `mapstructure:"urgency_id,omitempty" json:"urgency_id,omitempty"`
	Timezone   *string             `mapstructure:"timezone,omitempty" json:"timezone,omitempty"`
	Days       *[]string           `mapstructure:"days,omitempty" json:"days,omitempty"`
	TimeFrom   *string             `mapstructure:"time_from,omitempty" json:"time_from,omitempty"`
	TimeTo     *string             `mapstructure:"time_to,omitempty" json:"time_to,omitempty"`
	PolicyId   *int                `mapstructure:"policy_id,omitempty" json:"policy_id,omitempty"`
	Steps      *[]policyStepMember `mapstructure:"step_members" json:"step_members"`
}

type policy struct {
	Id            *int          `json:"id,omitempty"`
	Name          *string       `json:"name,omitempty"`
	RepeatCount   *int          `json:"repeat_count,omitempty"`
	RepeatDelay   *int          `json:"repeat_delay,omitempty"`
	IncidentToken *string       `json:"incident_token,omitempty"`
	Steps         *[]policyStep `json:"steps"`
}

type policyHTTPResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes policy `json:"attributes"`
	} `json:"data"`
}

func policyRef(in *policy) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "repeat_count", v: &in.RepeatCount},
		{k: "repeat_delay", v: &in.RepeatDelay},
		{k: "incident_token", v: &in.IncidentToken},
		{k: "steps", v: &in.Steps},
	}
}

func policyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in policy
	for _, e := range policyRef(&in) {
		if e.k == "steps" {
			loadPolicySteps(d, e.v.(**[]policyStep))
		} else {
			load(d, e.k, e.v)
		}
	}
	var out policyHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/policies", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return policyCopyAttrs(d, &out.Data.Attributes)
}

func policyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out policyHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/policies/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return policyCopyAttrs(d, &out.Data.Attributes)
}

func policyCopyAttrs(d *schema.ResourceData, in *policy) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range policyRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func policyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in policy
	var out policyHTTPResponse
	for _, e := range policyRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "steps" {
				loadPolicySteps(d, e.v.(**[]policyStep))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/policies/%s", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return policyCopyAttrs(d, &out.Data.Attributes)
}

func policyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/policies/%s", url.PathEscape(d.Id())))
}

func loadPolicySteps(d *schema.ResourceData, receiver **[]policyStep) {
	x := receiver
	stepsValues := d.Get("steps")

	var steps []policyStep

	for _, stepValues := range stepsValues.([]interface{}) {
		stepValuesObject := stepValues.(map[string]interface{})

		// Remove default values ("" or 0) from step
		for k, v := range stepValuesObject {
			if v == "" || v == 0 {
				stepValuesObject[k] = nil
			}
		}

		// Remove default values ("" or 0) from step members
		stepMembers, ok := stepValuesObject["step_members"].([]interface{})
		if ok {
			for _, member := range stepMembers {
				memberObject := member.(map[string]interface{})
				for k, v := range memberObject {
					if v == "" || v == 0 {
						memberObject[k] = nil
					}
				}
			}
		}

		var policyStep policyStep
		err := mapstructure.Decode(stepValuesObject, &policyStep)
		if err != nil {
			panic(err)
		}

		steps = append(steps, policyStep)
	}

	*x = &steps
}
