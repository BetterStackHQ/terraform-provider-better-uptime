package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var onCallCalendarSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     nil,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of the on-call calendar.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"name": {
		Description: "Name of the on-call calendar.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"default_calendar": {
		Description: "Whether the on-call calendar is the default on-call calendar.",
		Type:        schema.TypeBool,
		Optional:    false,
		Computed:    true,
	},
	"on_call_users": {
		Description: "Array of on-call persons.",
		Type:        schema.TypeList,
		Optional:    false,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Description: "ID of the on-call person.",
					Type:        schema.TypeString,
					Optional:    false,
					Computed:    true,
				},
				"first_name": {
					Description: "First name of the on-call person.",
					Type:        schema.TypeString,
					Optional:    false,
					Computed:    true,
				},
				"last_name": {
					Description: "Last name of the on-call person.",
					Type:        schema.TypeString,
					Optional:    false,
					Computed:    true,
				},
				"email": {
					Description: "Email of the on-call person.",
					Type:        schema.TypeString,
					Optional:    false,
					Computed:    true,
				},
				"phone_numbers": {
					Description: "Array of phone numbers.",
					Type:        schema.TypeList,
					Optional:    false,
					Computed:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	},
	"on_call_rotation": {
		Description: "TODO",
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"users": {
					Description: "TODO",
					Type:        schema.TypeList,
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"rotation_length": {
					Description: "TODO",
					Type:        schema.TypeInt,
					Required:    true,
				},
				"rotation_interval": {
					Description:  "TODO",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"hour", "day", "week"}, false),
				},
				"start_rotations_at": {
					Description:      "Start time of the rotation in RFC 3339 format (e.g. 2026-01-01T00:00:00Z)",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validateRFC3339DateTime,
					DiffSuppressFunc: diffSuppressRFC3339DateTime,
				},
				"end_rotations_at": {
					Description:      "End time of the rotation in RFC 3339 format (e.g. 2026-01-01T00:00:00Z)",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validateRFC3339DateTime,
					DiffSuppressFunc: diffSuppressRFC3339DateTime,
				},
			},
		},
	},
}

func newOnCallCalendarResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOnCallCalendarCreate,
		ReadContext:   resourceOnCallCalendarRead,
		UpdateContext: resourceOnCallCalendarUpdate,
		DeleteContext: resourceOnCallCalendarDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:      onCallCalendarSchema,
		Description: "https://betterstack.com/docs/uptime/api/on-call-calendar/",
	}
}

type onCallCalendar struct {
	ID              *string `json:"id,omitempty"`
	Name            *string `json:"name,omitempty"`
	DefaultCalendar *bool   `json:"default_calendar,omitempty"`
	TeamName        *string `json:"team_name,omitempty"`
}

type onCallRotation struct {
	Users            *[]string `mapstructure:"users,omitempty" json:"users,omitempty"`
	RotationLength   *int      `mapstructure:"rotation_length,omitempty" json:"rotation_length,omitempty"`
	RotationInterval *string   `mapstructure:"rotation_interval,omitempty" json:"rotation_interval,omitempty"`
	StartRotationsAt *string   `mapstructure:"start_rotations_at,omitempty" json:"start_rotations_at,omitempty"`
	EndRotationsAt   *string   `mapstructure:"end_rotations_at,omitempty" json:"end_rotations_at,omitempty"`
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

func onCallCalendarCopyAttrs(d *schema.ResourceData, cal *onCallCalendar, rel onCallRelationships, inc []onCallIncluded, rot *onCallRotation) diag.Diagnostics {
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

	// Only set rotation if it exists
	var rotationList []onCallRotation
	if rot != nil {
		rotationList = []onCallRotation{*rot}
	}
	if err := d.Set("on_call_rotation", rotationList); err != nil {
		derr = append(derr, diag.FromErr(err)[0])
	}

	return derr
}

func validateRFC3339DateTime(i interface{}, p cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type to be string")
	}

	if _, err := time.Parse(time.RFC3339, v); err != nil {
		return diag.Errorf("expected RFC 3339 datetime (e.g. 2026-01-01T00:00:00Z), got %s: %v", v, err)
	}

	return nil
}

func diffSuppressRFC3339DateTime(k, old, new string, d *schema.ResourceData) bool {
	if old == "" || new == "" {
		return false
	}

	oldTime, err := time.Parse(time.RFC3339, old)
	if err != nil {
		return false
	}

	newTime, err := time.Parse(time.RFC3339, new)
	if err != nil {
		return false
	}

	return oldTime.UTC().Equal(newTime.UTC())
}

func resourceOnCallCalendarCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in onCallCalendar
	for _, e := range onCallCalendarRef(&in) {
		load(d, e.k, e.v)
	}
	load(d, "team_name", &in.TeamName)

	var out struct {
		Data struct {
			ID            string              `json:"id"`
			Attributes    onCallCalendar      `json:"attributes"`
			Relationships onCallRelationships `json:"relationships"`
		} `json:"data"`
		Included []onCallIncluded `json:"included"`
	}

	if err := resourceCreate(ctx, meta, "/api/v2/on-calls", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)

	if inRotations, ok := d.GetOk("on_call_rotation"); ok && len(inRotations.([]interface{})) > 0 {
		var outRotation onCallRotation
		if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s/rotation", url.PathEscape(d.Id())), inRotations.([]interface{})[0], &outRotation); err != nil {
			return err
		}
		return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, &outRotation)
	}

	return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, nil)
}

func resourceOnCallCalendarRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out struct {
		Data struct {
			ID            string              `json:"id"`
			Attributes    onCallCalendar      `json:"attributes"`
			Relationships onCallRelationships `json:"relationships"`
		} `json:"data"`
		Included []onCallIncluded `json:"included"`
	}

	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404
		return nil
	}

	var outRotation onCallRotation
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s/rotation", url.PathEscape(d.Id())), &outRotation); err != nil {
		return err
	} else if !ok {
		return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, nil)
	}

	return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, &outRotation)
}

func resourceOnCallCalendarUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in onCallCalendar
	for _, e := range onCallCalendarRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}

	var out struct {
		Data struct {
			ID            string              `json:"id"`
			Attributes    onCallCalendar      `json:"attributes"`
			Relationships onCallRelationships `json:"relationships"`
		} `json:"data"`
		Included []onCallIncluded `json:"included"`
	}

	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}

	if d.HasChange("on_call_rotation") {
		if inRotations, ok := d.GetOk("on_call_rotation"); ok && len(inRotations.([]interface{})) > 0 {
			var outRotation onCallRotation
			if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s/rotation", url.PathEscape(d.Id())), inRotations.([]interface{})[0], &outRotation); err != nil {
				return err
			}
			return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, &outRotation)
		}
	}

	return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, nil)
}

func resourceOnCallCalendarDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s", url.PathEscape(d.Id())))
}
