package provider

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var onCallCalendarSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
	},
	"default_calendar": {
		Type:        schema.TypeBool,
		Computed:    true,
	},
	"on_call_users": {
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	},
}


type onCallCalendar struct {
	ID              *string      `json:"id,omitempty"`
	Name            *string      `json:"name,omitempty"`
	DefaultCalendar *bool        `json:"default_calendar,omitempty"`
	OnCallUsers     []onCallUser `json:"on_call_users,omitempty"`
}

type onCallRelationships struct {
	OnCallUsers struct {
		Data []onCallUser  `json:"data,omitempty"`
	} `json:"on_call_users,omitempty"`
}

type onCallUser struct {
	ID *string `mapstructure:"id,omitempty" json:"id,omitempty"`
}


func onCallCalendarRef(cal *onCallCalendar) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{"name", &cal.Name},
		{"default_calendar", &cal.DefaultCalendar},
	}
}

func onCallCalendarCopyAttrs(d *schema.ResourceData, cal *onCallCalendar, rel *onCallRelationships) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range onCallCalendarRef(cal) {
		if !isFieldAttribute(e.k) {
			value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
			if err := d.Set(e.k, value); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		}
	}
	if !isFieldAttribute("on_call_users") {
		value := reflect.Indirect(reflect.ValueOf(&rel.OnCallUsers.Data)).Interface()
		if err := d.Set("on_call_users", value); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}
