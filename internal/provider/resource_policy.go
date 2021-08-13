package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type policy struct {
	Name          *string   `json:"name,omitempty"`
	RepeatCount   *int      `json:"repeat_count,omitempty"`
	RepeatDelay   *int      `json:"repeat_delay,omitempty"`
	IncidentToken *string   `json:"incident_token,omitempty"`
}

type policyHTTPResponse struct {
	Data struct {
		ID         string  `json:"id"`
		Attributes policy `json:"attributes"`
	} `json:"data"`
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
		Type:     schema.TypeInt,
		Optional: true,
	},
	"repeat_delay": {
		Description: "", // TODO
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"incident_token": {
		Description: "", // TODO
		Type:        schema.TypeString,
		Required:    true,
	},
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
	}
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
