package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newPolicyDataSource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range policySchema {
		cp := *v
		switch k {
		case "name":
			// keep required
		default:
			cp.Computed = true
			cp.Optional = false
			cp.Required = false
			cp.ValidateDiagFunc = nil
			cp.Default = nil
			cp.DefaultFunc = nil
			cp.DiffSuppressFunc = nil
		}
		s[k] = &cp
	}
	return &schema.Resource{
		ReadContext: policyLookup,
		Description: "Policy lookup.",
		Schema:      s,
	}
}

type policy struct {
	Name          *string `json:"name,omitempty"`
	RepeatCount   *int    `json:"repeat_count,omitempty"`
	RepeatDelay   *int    `json:"repeat_delay,omitempty"`
	IncidentToken *string `json:"incident_token,omitempty"`
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

func policyCopyAttrs(d *schema.ResourceData, in *policy) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range policyRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

type policyPageHTTPResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes policy `json:"attributes"`
	} `json:"data"`
	Pagination struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"pagination"`
}

func policyLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fetch := func(page int) (*policyPageHTTPResponse, error) {
		res, err := meta.(*client).Get(ctx, fmt.Sprintf("/api/v2/policies?page=%d", page))
		if err != nil {
			return nil, err
		}
		defer func() {
			// Keep-Alive.
			_, _ = io.Copy(ioutil.Discard, res.Body)
			_ = res.Body.Close()
		}()
		body, err := ioutil.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
		}
		if err != nil {
			return nil, err
		}
		var tr policyPageHTTPResponse
		return &tr, json.Unmarshal(body, &tr)
	}
	name := d.Get("name").(string)
	page := 1
	for {
		res, err := fetch(page)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, e := range res.Data {
			if *e.Attributes.Name == name {
				if d.Id() != "" {
					return diag.Errorf("duplicate")
				}
				d.SetId(e.ID)
				if derr := policyCopyAttrs(d, &e.Attributes); derr != nil {
					return derr
				}
			}
		}
		page++
		if res.Pagination.Next == "" {
			return nil
		}
	}
}
