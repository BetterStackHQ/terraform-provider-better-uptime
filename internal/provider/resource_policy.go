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
		Description: "", // TODO
		Type:        schema.TypeString,
		Required:    true,
	},
	"id": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"team_id": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
	},
}

var policyStepSchema = map[string]*schema.Schema{
	"type": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Required:    true,
	},
	"wait_before": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Required:    true,
	},
	"urgency_id": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"timezone": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"days": {
		Description: "", // TODO
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"time_from": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"time_to": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Optional:    true,
	},
	"policy_id": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"step_members": {
		Description: "An array of escalation policy steps members", // TODO expand
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
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
	"repeat_delay": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
	"incident_token": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Computed:    true,
	},
	"steps": {
		Description: "An array of escalation policy steps", // TODO expand
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
		Description: "https://docs.betteruptime.com/api/policies-api",
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
	for _, e := range policyRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "steps" {
				loadPolicySteps(d, e.v.(**[]policyStep))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/policies/%s", url.PathEscape(d.Id())), &in)
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
		var policyStep policyStep
		err := mapstructure.Decode(stepValuesObject, &policyStep)
		if err != nil {
			panic(err)
		}

		steps = append(steps, policyStep)
	}

	*x = &steps
}
