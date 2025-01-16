package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var attributeSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the attribute.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"primary": {
		Description: "Whether this attribute is the primary attribute.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"position": {
		Description: "The position of the attribute in the catalog relation.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
	},
}

var catalogRelationSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the catalog relation.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"description": {
		Description: "A description of the catalog relation.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"attributes": {
		Description: "Attributes of the catalog relation.",
		Type:        schema.TypeList,
		MinItems:    1,
		Elem:        &schema.Resource{Schema: attributeSchema},
		Required:    true,
	},
	"records": {
		Description:      "Records of the catalog relation.",
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: suppressEquivalentJSONDiffs,
	},
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

func catalogRelationAttributeRef(in *catalogRelationAttribute) []struct {
	k string
	v interface{}
} {
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "primary", v: &in.Primary},
		{k: "position", v: &in.Position},
	}
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
		Description: "Resource for managing Better Uptime catalog relations.",
		Schema:      catalogRelationSchema,
	}
}

type attribute struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	Required    *bool   `json:"required,omitempty"`
	Position    *int    `json:"position,omitempty"`
}

type catalogRelation struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type catalogRelationHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes catalogRelation `json:"attributes"`
	} `json:"data"`
}

type catalogRelationAttribute struct {
	Name     *string `json:"name,omitempty"`
	Primary  *bool   `json:"primary,omitempty"`
	Position *int    `json:"position,omitempty"`
}

type catalogRelationAttributesHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes catalogRelation `json:"attributes"`
	} `json:"data"`
}

func catalogRelationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Create the relation
	if diags := createRelation(ctx, d, meta); diags != nil {
		return diags
	}

	// Create the attributes
	attributes := d.Get("attributes").([]interface{})
	for _, attribute := range attributes {
		if diags := createRelationAttribute(ctx, attribute, meta); diags != nil {
			return diags
		}
	}

	return nil
}

func createRelation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in catalogRelation

	for _, v := range catalogRelationRef(&in) {
		load(d, v.k, v.v)
	}

	var out catalogRelationHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/catalog/relations", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)

	return catalogRelationCopyAttrs(d, &out.Data.Attributes)
}

func createRelationAttribute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in catalogRelationAttribute
	for _, e := range catalogRelationAttributeRef(&in) {
		load(d, e.k, e.v)
	}

	var out catalogRelationAttributesHTTPResponse
	if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/catalog/relations/%s/attributes", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}

	return nil
}

func catalogRelationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// var out catalogRelationHTTPResponse
	// if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/catalog-relations/%s", url.PathEscape(d.Id())), &out); err != nil {
	// 	return err
	// } else if !ok {
	// 	d.SetId("") // Force "create" on 404.
	// 	return nil
	// }
	// return catalogRelationCopyAttrs(d, &out.Data.Attributes)
	return nil
}

func catalogRelationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// var in catalogRelation
	// var out catalogRelationHTTPResponse

	// // Update only the changed fields
	// if d.HasChange("name") {
	// 	load(d, "name", &in.Name)
	// }
	// if d.HasChange("description") {
	// 	load(d, "description", &in.Description)
	// }
	// if d.HasChange("attributes") {
	// 	loadAttributes(d, "attributes", &in.Attributes)
	// }
	// if d.HasChange("records") {
	// 	loadRecords(d, "records", &in.Records)
	// }

	// return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/catalog-relations/%s", url.PathEscape(d.Id())), &in, &out)
	return nil
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

func catalogRelationAttributeCopyAttrs(d *schema.ResourceData, in *catalogRelationAttribute) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range catalogRelationAttributeRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}
