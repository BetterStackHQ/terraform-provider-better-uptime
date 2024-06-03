package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var pagerdutyIntegrationSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of the PagerDuty Integration.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "The name of the PagerDuty Integration.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"key": {
		Description: "The PagerDuty routing key.",
		Type:        schema.TypeString,
		Required:    true,
	},
}

func newPagerdutyIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: pagerdutyIntegrationCreate,
		ReadContext:   pagerdutyIntegrationRead,
		UpdateContext: pagerdutyIntegrationUpdate,
		DeleteContext: pagerdutyIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/pagerduty-integrations/",
		Schema:      pagerdutyIntegrationSchema,
	}
}

type pagerdutyIntegration struct {
	ID       *string `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Key      *string `json:"key,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
}

type pagerdutyIntegrationHTTPResponse struct {
	Data struct {
		ID         string               `json:"id"`
		Attributes pagerdutyIntegration `json:"attributes"`
	} `json:"data"`
}

func pagerdutyIntegrationRef(in *pagerdutyIntegration) []struct {
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
		{k: "key", v: &in.Key},
	}
}

func pagerdutyIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in pagerdutyIntegration
	for _, e := range pagerdutyIntegrationRef(&in) {
		if e.k == "request_headers" {
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
	var out pagerdutyIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/pager-duty-webhooks", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return pagerdutyIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func pagerdutyIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out pagerdutyIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/pager-duty-webhooks/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return pagerdutyIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func pagerdutyIntegrationCopyAttrs(d *schema.ResourceData, in *pagerdutyIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range pagerdutyIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func pagerdutyIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in pagerdutyIntegration
	var out pagerdutyIntegrationHTTPResponse
	for _, e := range pagerdutyIntegrationRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/pager-duty-webhooks/%s", url.PathEscape(d.Id())), &in, &out)
}

func pagerdutyIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/pager-duty-webhooks/%s", url.PathEscape(d.Id())))
}
