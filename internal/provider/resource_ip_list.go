package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ipListSchema = map[string]*schema.Schema{
	"id": {
		Description: "The internal ID of the resource, can be ignored.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"ips": {
		Description: "The list of IPs used for monitoring.",
		Type:        schema.TypeList,
		Optional:    false,
		Computed:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"all_clusters": {
		Description: "The list of all clusters.",
		Type:        schema.TypeList,
		Optional:    false,
		Computed:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"filter_clusters": {
		Description: "The optional list of clusters used to filter the IPs.",
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    false,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
}
