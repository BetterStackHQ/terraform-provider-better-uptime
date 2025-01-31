package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var catalogRelationSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Catalog relation.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "The name of the Catalog relation.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"description": {
		Description: "A description of the Catalog relation.",
		Type:        schema.TypeString,
		Optional:    true,
	},
}

func newCatalogRelationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: catalogRelationCreate,
		ReadContext:   catalogRelationRead,
		UpdateContext: catalogRelationUpdate,
		DeleteContext: catalogRelationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/catalog-integrations-relations/",
		Schema:      catalogRelationSchema,
	}
}

type catalogRelation struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type catalogRelationHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes catalogRelation `json:"attributes"`
	} `json:"data"`
}

func catalogRelationRef(in *catalogRelation) []struct {
	k string
	v interface{}
} {
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "description", v: &in.Description},
	}
}

func catalogRelationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in catalogRelation
	for _, e := range catalogRelationRef(&in) {
		load(d, e.k, e.v)
	}

	var out catalogRelationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/catalog/relations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return catalogRelationCopyAttrs(d, &out.Data.Attributes)
}

func catalogRelationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out catalogRelationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("")
		return nil
	}
	return catalogRelationCopyAttrs(d, &out.Data.Attributes)
}

func catalogRelationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in catalogRelation
	var out catalogRelationHTTPResponse
	for _, e := range catalogRelationRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}

	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	return catalogRelationCopyAttrs(d, &out.Data.Attributes)
}

func catalogRelationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s", url.PathEscape(d.Id())))
}

func catalogRelationCopyAttrs(d *schema.ResourceData, in *catalogRelation) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range catalogRelationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}
