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
	case **statusHistory:
		if v, ok := d.GetOkExists(key); ok {
			t := v.(statusHistory)
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
