package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// nolint
func load(d *schema.ResourceData, key string, receiver interface{}) {
	switch x := receiver.(type) {
	case **string:
		if v, ok := d.GetOkExists(key); ok {
			t := v.(string)
			*x = &t
		}
	case **int:
		if v, ok := d.GetOkExists(key); ok {
			t := v.(int)
			*x = &t
		}
	case **float32:
		if v, ok := d.GetOkExists(key); ok {
			t := v.(float32)
			*x = &t
		}
	case **bool:
		if v, ok := d.GetOkExists(key); ok {
			t := v.(bool)
			*x = &t
		}
	case **[]string:
		if v, ok := d.GetOkExists(key); ok {
			var t = make([]string, 0)
			for _, v := range v.([]interface{}) {
				t = append(t, v.(string))
			}
			*x = &t
		}
	case **[]int:
		if v, ok := d.GetOkExists(key); ok {
			var t = make([]int, 0)
			for _, v := range v.([]interface{}) {
				t = append(t, v.(int))
			}
			*x = &t
		}
	case **map[string]string:
		if v, ok := d.GetOkExists(key); ok {
			var t = make(map[string]string, 0)
			for k, v := range v.(map[string]interface{}) {
				t[k] = v.(string)
			}
			*x = &t
		}
	default:
		panic(fmt.Errorf("unexpected type %T", receiver))
	}
}

func truePtr() *bool {
	b := true
	return &b
}

// dropUnknownKeys removes keys not defined in the schema from raw API maps,
// so fields newly added to the API are ignored instead of failing d.Set.
func dropUnknownKeys(in *[]map[string]interface{}, s map[string]*schema.Schema) *[]map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make([]map[string]interface{}, len(*in))
	for i, m := range *in {
		filtered := make(map[string]interface{}, len(m))
		for k, v := range m {
			if _, ok := s[k]; ok {
				filtered[k] = v
			}
		}
		out[i] = filtered
	}
	return &out
}
