package provider

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var onCallCalendarSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of the on-call calendar.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "Name of the on-call calendar.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"default_calendar": {
		Description: "Whether the on-call calendar is the default on-call calendar.",
		Type:        schema.TypeBool,
		Computed:    true,
	},
	"on_call_users": {
		Description: "Array of on-call persons.",
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Description: "ID of the on-call person.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"first_name": {
					Description: "First name of the on-call person.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"last_name": {
					Description: "Last name of the on-call person.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"email": {
					Description: "Email of the on-call person.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"phone_numbers": {
					Description: "Array of phone numbers.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	},
}

type onCallCalendar struct {
	ID              *string `json:"id,omitempty"`
	Name            *string `json:"name,omitempty"`
	DefaultCalendar *bool   `json:"default_calendar,omitempty"`
	TeamName        *string `json:"team_name,omitempty"`
}

type onCallRelationships struct {
	OnCallUsers struct {
		Data []struct {
			ID           *string   `mapstructure:"id,omitempty" json:"id,omitempty"`
			FirstName    *string   `mapstructure:"first_name,omitempty"`
			LastName     *string   `mapstructure:"last_name,omitempty"`
			Email        *string   `mapstructure:"email,omitempty"`
			PhoneNumbers []*string `mapstructure:"phone_numbers,omitempty"`
		} `json:"data,omitempty"`
	} `json:"on_call_users,omitempty"`
}

type onCallIncluded struct {
	ID         *string `json:"id,omitempty"`
	Type       *string `json:"type,omitempty"`
	Attributes struct {
		FirstName    *string   `json:"first_name,omitempty"`
		LastName     *string   `json:"last_name,omitempty"`
		Email        *string   `json:"email,omitempty"`
		PhoneNumbers []*string `json:"phone_numbers,omitempty"`
	} `json:"attributes,omitempty"`
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

func onCallCalendarCopyAttrs(d *schema.ResourceData, cal *onCallCalendar, rel onCallRelationships, inc []onCallIncluded) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range onCallCalendarRef(cal) {
		if !isFieldAttribute(e.k) {
			value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
			if err := d.Set(e.k, value); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		}
	}
	// Enrich relationships data from included values
	for i := range rel.OnCallUsers.Data {
		for _, e := range inc {
			if *rel.OnCallUsers.Data[i].ID == *e.ID && *e.Type == "user" {
				rel.OnCallUsers.Data[i].FirstName = e.Attributes.FirstName
				rel.OnCallUsers.Data[i].LastName = e.Attributes.LastName
				rel.OnCallUsers.Data[i].Email = e.Attributes.Email
				rel.OnCallUsers.Data[i].PhoneNumbers = e.Attributes.PhoneNumbers
			}
		}
	}
	if !isFieldAttribute("on_call_users") {
		if err := d.Set("on_call_users", rel.OnCallUsers.Data); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}
