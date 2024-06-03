package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var splunkOnCallIntegrationSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of the Splunk On-Call Integration.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "The name of the Splunk On-Call Integration.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"url": {
		Description: "The Splunk On-Call URL to post webhooks to.",
		Type:        schema.TypeString,
		Required:    true,
	},
}

func newSplunkOnCallIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: splunkOnCallIntegrationCreate,
		ReadContext:   splunkOnCallIntegrationRead,
		UpdateContext: splunkOnCallIntegrationUpdate,
		DeleteContext: splunkOnCallIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/splunk-on-call-integrations/",
		Schema:      splunkOnCallIntegrationSchema,
	}
}

type splunkOnCallIntegration struct {
	ID       *string `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	URL      *string `json:"url,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
}

type splunkOnCallIntegrationHTTPResponse struct {
	Data struct {
		ID         string                  `json:"id"`
		Attributes splunkOnCallIntegration `json:"attributes"`
	} `json:"data"`
}

func splunkOnCallIntegrationRef(in *splunkOnCallIntegration) []struct {
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
		{k: "url", v: &in.URL},
		{k: "team_name", v: &in.TeamName},
	}
}

func splunkOnCallIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in splunkOnCallIntegration
	for _, e := range splunkOnCallIntegrationRef(&in) {
		if e.k == "request_headers" {
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
	var out splunkOnCallIntegrationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/splunk-on-calls", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return splunkOnCallIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func splunkOnCallIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out splunkOnCallIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/splunk-on-calls/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return splunkOnCallIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func splunkOnCallIntegrationCopyAttrs(d *schema.ResourceData, in *splunkOnCallIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range splunkOnCallIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func splunkOnCallIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in splunkOnCallIntegration
	var out splunkOnCallIntegrationHTTPResponse
	for _, e := range splunkOnCallIntegrationRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/splunk-on-calls/%s", url.PathEscape(d.Id())), &in, &out)
}

func splunkOnCallIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/splunk-on-calls/%s", url.PathEscape(d.Id())))
}
