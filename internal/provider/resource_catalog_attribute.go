package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var catalogAttributeSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Catalog attribute.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"relation_id": {
		Description: "The ID of the Catalog relation this attribute belongs to.",
		Type:        schema.TypeString,
		ForceNew:    true,
		Required:    true,
	},
	"name": {
		Description: "The name of the Catalog attribute.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"primary": {
		Description: "Whether this attribute is one of the primary attributes.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"position": {
		Description: "The position of the attribute in the Catalog relation.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
}

func newCatalogAttributeResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: catalogAttributeCreate,
		ReadContext:   catalogAttributeRead,
		UpdateContext: catalogAttributeUpdate,
		DeleteContext: catalogAttributeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				split := strings.SplitN(d.Id(), "/", 2)
				if len(split) != 2 {
					return nil, errors.New("betteruptime_catalog_attribute can be imported via \"relation_id/id\" only (e.g. \"123/234\")")
				}
				if err := d.Set("relation_id", split[0]); err != nil {
					return nil, err
				}
				d.SetId(split[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: "https://betterstack.com/docs/uptime/api/create-catalog-attribute/",
		Schema:      catalogAttributeSchema,
	}
}

type catalogAttribute struct {
	ID      *string `json:"id,omitempty"`
	Name    *string `json:"name,omitempty"`
	Primary *bool   `json:"primary,omitempty"`
}

type catalogAttributeHTTPResponse struct {
	Data struct {
		ID         string           `json:"id"`
		Attributes catalogAttribute `json:"attributes"`
	} `json:"data"`
}

func catalogAttributeRef(in *catalogAttribute) []struct {
	k string
	v interface{}
} {
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "primary", v: &in.Primary},
	}
}

func catalogAttributeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in catalogAttribute
	for _, e := range catalogAttributeRef(&in) {
		load(d, e.k, e.v)
	}

	relationID := d.Get("relation_id").(string)
	var out catalogAttributeHTTPResponse
	if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/attributes", url.PathEscape(relationID)), &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return catalogAttributeCopyAttrs(d, &out.Data.Attributes)
}

func catalogAttributeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relationID := d.Get("relation_id").(string)
	var out catalogAttributeHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/attributes/%s", url.PathEscape(relationID), url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("")
		return nil
	}
	return catalogAttributeCopyAttrs(d, &out.Data.Attributes)
}

func catalogAttributeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in catalogAttribute
	var out catalogAttributeHTTPResponse
	for _, e := range catalogAttributeRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}

	relationID := d.Get("relation_id").(string)
	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/attributes/%s", url.PathEscape(relationID), url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	return catalogAttributeCopyAttrs(d, &out.Data.Attributes)
}

func catalogAttributeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	relationID := d.Get("relation_id").(string)
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/attributes/%s", url.PathEscape(relationID), url.PathEscape(d.Id())))
}

func catalogAttributeCopyAttrs(d *schema.ResourceData, in *catalogAttribute) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range catalogAttributeRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}
