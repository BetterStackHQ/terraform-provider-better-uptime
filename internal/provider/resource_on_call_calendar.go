package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var onCallCalendarSchema = map[string]*schema.Schema{
	"team_name": teamNameSchema(),
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
	// Optional and computed, so a calendar tagged in Better Stack plans no change while the configuration leaves it out.
	"slack_group_handle": {
		Description: "The handle of the Slack user group that tags whoever is on call in this calendar, without the `@`: lowercase letters, digits, `.`, `_` and `-`. Better Stack keeps the group's members in sync with the rotation. Set it to an empty string to stop tagging; leaving it out keeps the handle the calendar has.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"use_existing_slack_group": {
		Description: "Whether `slack_group_handle` names a user group that already exists in Slack. `true` uses that group: Better Stack adds and removes only the people on call and keeps its other members. `false` creates a new group whose members are exactly the people on call, and a handle an existing group holds is refused. Leaving it out keeps the calendar's current choice (a new group if it has none).",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"slack_groups": {
		Description: "How each connected Slack workspace took the handle. Better Stack sets the groups up in the background after a change: an apply that changes the calendar waits for that (up to 5 minutes), then warns about any workspace where the group doesn't work and about responders it can't tag.",
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"slack_workspace_id": {
					Description: "The ID of the Slack workspace.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"status": {
					Description: "`active` once the group tags whoever is on call, `pending` while it is set up. `reinstall_required` when the workspace was connected before Better Stack could manage user groups, and any other status when Slack refused the group, with the reason in `error`.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"error": {
					Description: "Why the group doesn't work in this workspace, and what to do about it.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"unconnected_users": {
					Description: "Responders the group can't tag because they haven't connected Slack in this workspace. They can connect by using the same e-mail in Slack and Better Stack, or in their account settings.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Description: "ID of the responder.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"first_name": {
								Description: "First name of the responder.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"last_name": {
								Description: "Last name of the responder.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"email": {
								Description: "E-mail of the responder.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
			},
		},
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
		Description: "Configuration block for the on-call rotation schedule. Ignored when omitted - on-call can be controlled in Better Stack.",
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"users": {
					Description: "List of email addresses for users participating in the rotation.",
					Type:        schema.TypeList,
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"rotation_length": {
					Description: "The length of each rotation shift. See `rotation_interval` for units.",
					Type:        schema.TypeInt,
					Required:    true,
				},
				"rotation_interval": {
					Description:  "The interval unit for rotation_length. Must be one of: `hour`, `day`, `week`.",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"hour", "day", "week"}, false),
				},
				"start_rotations_at": {
					Description:      "Start time of the rotation in RFC 3339 format (e.g. `2026-01-01T00:00:00Z`)",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validateRFC3339DateTime,
					DiffSuppressFunc: diffSuppressRFC3339DateTime,
				},
				"end_rotations_at": {
					Description:      "End time of the rotation in RFC 3339 format (e.g. `2026-01-01T00:00:00Z`)",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validateRFC3339DateTime,
					DiffSuppressFunc: diffSuppressRFC3339DateTime,
				},
				"timezone": {
					Description:  "Time zone the rotation's working hours are written in and its handovers follow across daylight-saving changes, an IANA name such as `Europe/Prague`. The offset in `start_rotations_at` still fixes the instant of the first shift, so write it in this time zone's offset. Omit to keep the time zone the rotation already has; a new rotation with working hours and no time zone uses UTC.",
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				"working_hours": {
					Description: "Windows during which the rotation pages anyone, one block per day and window. Omit for a rotation that is active around the clock; omitting it on an existing rotation removes its working hours. A window whose `end_time` is not after `start_time` runs overnight.",
					Type:        schema.TypeList,
					Optional:    true,
					// The API rejects a 51st window, and a plan-time error beats an apply that fails halfway.
					MaxItems: 50,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"day": {
								Description:  "Day of the week. Must be one of: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`.",
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(onCallWeekDays, false),
							},
							"start_time": {
								Description:      "Start of the window as `HH:MM` on a 24-hour clock (e.g. `09:00`). Starts at midnight when omitted.",
								Type:             schema.TypeString,
								Optional:         true,
								Default:          "00:00",
								ValidateDiagFunc: validateTimeOfDay,
								DiffSuppressFunc: diffSuppressTimeOfDay,
							},
							"end_time": {
								Description:      "End of the window as `HH:MM` on a 24-hour clock (e.g. `17:00`). Ends at midnight when omitted, so a block with both times omitted covers the whole day.",
								Type:             schema.TypeString,
								Optional:         true,
								Default:          "00:00",
								ValidateDiagFunc: validateTimeOfDay,
								DiffSuppressFunc: diffSuppressTimeOfDay,
							},
						},
					},
				},
			},
		},
	},
}

var onCallWeekDays = []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

// HH:MM on a 24-hour clock, with an optional :00 for callers that write seconds; the API rejects any other seconds.
var timeOfDayPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9](:00)?$`)

func newOnCallCalendarResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOnCallCalendarCreate,
		ReadContext:   resourceOnCallCalendarRead,
		UpdateContext: resourceOnCallCalendarUpdate,
		DeleteContext: resourceOnCallCalendarDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			validateTeamNameNotChanged,
			planSlackGroupHandleCleared,
			// A changed handle or mode is set up again in the background, so how each workspace takes it is only known after
			// the apply. Not on_call_rotation, which reads as changed on every plan here; who the group can't tag still
			// follows the rotation, set by the apply.
			customdiff.ComputedIf("slack_groups", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChanges("slack_group_handle", "use_existing_slack_group")
			}),
		),
		Schema:      onCallCalendarSchema,
		Description: "https://betterstack.com/docs/uptime/api/on-call-calendar/",
	}
}

// An optional and computed string reads an empty one as left out, which keeps the old value. Only the raw configuration
// tells a configured "" (stop tagging) from an absent attribute (keep the handle).
func planSlackGroupHandleCleared(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if old, _ := d.GetChange("slack_group_handle"); old.(string) == "" || !slackGroupHandleConfiguredEmpty(d.GetRawConfig()) {
		return nil
	}
	return d.SetNew("slack_group_handle", "")
}

func slackGroupHandleConfiguredEmpty(raw cty.Value) bool {
	if raw.IsNull() || !raw.IsKnown() {
		return false
	}
	handle := raw.GetAttr("slack_group_handle")
	return handle.IsKnown() && !handle.IsNull() && handle.AsString() == ""
}

type onCallCalendar struct {
	ID                    *string `json:"id,omitempty"`
	Name                  *string `json:"name,omitempty"`
	DefaultCalendar       *bool   `json:"default_calendar,omitempty"`
	TeamName              *string `json:"team_name,omitempty"`
	SlackGroupHandle      *string `json:"slack_group_handle,omitempty"`
	UseExistingSlackGroup *bool   `json:"use_existing_slack_group,omitempty"`
	// Read-only, so never loaded from the configuration.
	SlackGroups *[]onCallSlackGroup `json:"slack_groups,omitempty"`
}

type onCallSlackGroup struct {
	SlackWorkspaceID *string `json:"slack_workspace_id,omitempty"`
	Status           *string `json:"status,omitempty"`
	Error            *string `json:"error,omitempty"`
	// Only on the calendar's own GET, not in a write's response.
	UnconnectedUsers []onCallSlackUser `json:"unconnected_users,omitempty"`
}

type onCallSlackUser struct {
	ID        *string `json:"id,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
}

type onCallWorkingHour struct {
	Day       *string `mapstructure:"day,omitempty" json:"day,omitempty"`
	StartTime *string `mapstructure:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime   *string `mapstructure:"end_time,omitempty" json:"end_time,omitempty"`
}

type onCallRotation struct {
	Users            *[]string            `mapstructure:"users,omitempty" json:"users,omitempty"`
	RotationLength   *int                 `mapstructure:"rotation_length,omitempty" json:"rotation_length,omitempty"`
	RotationInterval *string              `mapstructure:"rotation_interval,omitempty" json:"rotation_interval,omitempty"`
	StartRotationsAt *string              `mapstructure:"start_rotations_at,omitempty" json:"start_rotations_at,omitempty"`
	EndRotationsAt   *string              `mapstructure:"end_rotations_at,omitempty" json:"end_rotations_at,omitempty"`
	Timezone         *string              `mapstructure:"timezone,omitempty" json:"timezone,omitempty"`
	WorkingHours     *[]onCallWorkingHour `mapstructure:"working_hours,omitempty" json:"working_hours,omitempty"`
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
		{"slack_group_handle", &cal.SlackGroupHandle},
		{"use_existing_slack_group", &cal.UseExistingSlackGroup},
	}
}

func onCallCalendarCopyAttrs(d *schema.ResourceData, cal *onCallCalendar, rel onCallRelationships, inc []onCallIncluded, rot *onCallRotation) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range onCallCalendarRef(cal) {
		value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
		if err := d.Set(e.k, value); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
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
	slackGroups := []map[string]interface{}{}
	if cal.SlackGroups != nil {
		for _, group := range *cal.SlackGroups {
			unconnected := []map[string]interface{}{}
			for _, user := range group.UnconnectedUsers {
				unconnected = append(unconnected, map[string]interface{}{
					"id":         stringValue(user.ID),
					"first_name": stringValue(user.FirstName),
					"last_name":  stringValue(user.LastName),
					"email":      stringValue(user.Email),
				})
			}
			slackGroups = append(slackGroups, map[string]interface{}{
				"slack_workspace_id": stringValue(group.SlackWorkspaceID),
				"status":             stringValue(group.Status),
				"error":              stringValue(group.Error),
				"unconnected_users":  unconnected,
			})
		}
	}
	if err := d.Set("slack_groups", slackGroups); err != nil {
		derr = append(derr, diag.FromErr(err)[0])
	}
	if err := d.Set("on_call_users", rel.OnCallUsers.Data); err != nil {
		derr = append(derr, diag.FromErr(err)[0])
	}

	// Only set rotation if it exists
	var rotationList []onCallRotation
	if rot != nil {
		rotation := *rot
		if rotation.WorkingHours != nil {
			workingHours := reorderOnCallWorkingHours(d, *rotation.WorkingHours)
			rotation.WorkingHours = &workingHours
		}
		rotationList = []onCallRotation{rotation}
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

func validateTimeOfDay(i interface{}, p cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type to be string")
	}

	if !timeOfDayPattern.MatchString(v) {
		return diag.Errorf("expected a time of day as HH:MM (24-hour clock), got %s", v)
	}

	return nil
}

func diffSuppressTimeOfDay(k, old, new string, d *schema.ResourceData) bool {
	return normalizeTimeOfDay(old) == normalizeTimeOfDay(new)
}

// Seconds are dropped rather than compared: the API stores HH:MM:SS but only ever returns HH:MM.
func normalizeTimeOfDay(value string) string {
	if len(value) > 5 && timeOfDayPattern.MatchString(value) {
		return value[:5]
	}

	return value
}

// The API stores the windows as a set and returns them in canonical order, Sunday first and then by
// start time; an ordered block list has to keep the operator's order or every plan diffs on order alone.
func reorderOnCallWorkingHours(d *schema.ResourceData, workingHours []onCallWorkingHour) []onCallWorkingHour {
	existing, ok := d.Get("on_call_rotation.0.working_hours").([]interface{})
	if !ok || len(existing) == 0 {
		return workingHours
	}

	value := func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}

	used := make([]bool, len(workingHours))
	ordered := make([]onCallWorkingHour, 0, len(workingHours))

	for _, e := range existing {
		em, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		day, _ := em["day"].(string)
		startTime, _ := em["start_time"].(string)
		endTime, _ := em["end_time"].(string)
		signature := onCallWorkingHourSignature(day, startTime, endTime)

		for i, workingHour := range workingHours {
			if used[i] {
				continue
			}
			if onCallWorkingHourSignature(value(workingHour.Day), value(workingHour.StartTime), value(workingHour.EndTime)) == signature {
				used[i] = true
				ordered = append(ordered, workingHour)
				break
			}
		}
	}

	// A window the configuration does not have, such as one added in Better Stack, lands at the end in
	// the API's own order, so the next plan still shows it as a removal.
	for i, workingHour := range workingHours {
		if !used[i] {
			ordered = append(ordered, workingHour)
		}
	}

	return ordered
}

// The times are compared as HH:MM, so a configured 09:00:00 pairs with the 09:00 the API returns.
func onCallWorkingHourSignature(day, startTime, endTime string) string {
	return strings.Join([]string{day, normalizeTimeOfDay(startTime), normalizeTimeOfDay(endTime)}, "\x00")
}

// Terraform hands back every key of the block, zero-valued when unconfigured; the API treats a blank
// timezone and an empty working_hours list as "not sent", and the request should say the same.
func onCallRotationRequest(block interface{}) map[string]interface{} {
	request := map[string]interface{}{}
	for key, value := range block.(map[string]interface{}) {
		request[key] = value
	}

	if timezone, _ := request["timezone"].(string); timezone == "" {
		delete(request, "timezone")
	}
	if workingHours, _ := request["working_hours"].([]interface{}); len(workingHours) == 0 {
		delete(request, "working_hours")
	}

	return request
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

	var rotation *onCallRotation
	if inRotations, ok := d.GetOk("on_call_rotation"); ok && len(inRotations.([]interface{})) > 0 {
		var outRotation onCallRotation
		if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s/rotation", url.PathEscape(d.Id())), onCallRotationRequest(inRotations.([]interface{})[0]), &outRotation); err != nil {
			return err
		}
		rotation = &outRotation
	} else {
		var derr diag.Diagnostics
		if rotation, derr = readOnCallRotation(ctx, meta, d.Id()); derr != nil {
			return derr
		}
	}

	return onCallCalendarApplied(ctx, d, meta, &out.Data.Attributes, out.Data.Relationships, out.Included, rotation)
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
	if derr := setTeamNameFromAPI(d, out.Data.Attributes.TeamName); derr != nil {
		return derr
	}

	rotation, derr := readOnCallRotation(ctx, meta, d.Id())
	if derr != nil {
		return derr
	}

	return onCallCalendarCopyAttrs(d, &out.Data.Attributes, out.Data.Relationships, out.Included, rotation)
}

// Nil when the calendar has no rotation.
func readOnCallRotation(ctx context.Context, meta interface{}, id string) (*onCallRotation, diag.Diagnostics) {
	var rotation onCallRotation
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s/rotation", url.PathEscape(id)), &rotation); err != nil || !ok {
		return nil, err
	}
	return &rotation, nil
}

// How long an apply waits for Better Stack to set up the calendar's Slack user groups, which usually takes seconds, and
// how often it checks. Variables so tests can shorten them.
var (
	slackGroupsSetupWait         = 5 * time.Minute
	slackGroupsSetupPollInterval = 3 * time.Second
)

// Copies a written calendar into the state once its Slack user groups are set up, and warns about what keeps them from
// tagging whoever is on call, as the editor in Better Stack does. Only an apply waits and warns: a refresh only reads.
func onCallCalendarApplied(ctx context.Context, d *schema.ResourceData, meta interface{}, cal *onCallCalendar, rel onCallRelationships, inc []onCallIncluded, rot *onCallRotation) diag.Diagnostics {
	var warnings diag.Diagnostics
	if handle := stringValue(cal.SlackGroupHandle); handle != "" {
		groups, derr := awaitSlackGroups(ctx, meta, d.Id())
		if derr != nil {
			return derr
		}
		cal.SlackGroups = groups
		warnings = slackGroupWarnings(handle, groups)
	}
	return append(onCallCalendarCopyAttrs(d, cal, rel, inc, rot), warnings...)
}

// Reads the calendar until no workspace is pending, or slackGroupsSetupWait passes. The calendar's own GET, unlike a
// write's response, also names the responders each group can't tag.
func awaitSlackGroups(ctx context.Context, meta interface{}, id string) (*[]onCallSlackGroup, diag.Diagnostics) {
	deadline := time.Now().Add(slackGroupsSetupWait)
	for {
		groups, derr := readSlackGroups(ctx, meta, id)
		if derr != nil {
			return nil, derr
		}
		if !slackGroupsPending(groups) || !time.Now().Add(slackGroupsSetupPollInterval).Before(deadline) {
			return groups, nil
		}
		select {
		case <-ctx.Done():
			return groups, nil
		case <-time.After(slackGroupsSetupPollInterval):
		}
	}
}

// The calendar's slack_groups as its own GET reports them, with the responders each group can't tag, which a list leaves out.
func readSlackGroups(ctx context.Context, meta interface{}, id string) (*[]onCallSlackGroup, diag.Diagnostics) {
	var out struct {
		Data struct {
			Attributes onCallCalendar `json:"attributes"`
		} `json:"data"`
	}
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s", url.PathEscape(id)), &out); err != nil {
		return nil, err
	} else if !ok {
		return nil, diag.Errorf("on-call calendar %s was not found", id)
	}
	return out.Data.Attributes.SlackGroups, nil
}

func slackGroupsPending(groups *[]onCallSlackGroup) bool {
	if groups == nil {
		return false
	}
	for _, group := range *groups {
		if stringValue(group.Status) == "pending" {
			return true
		}
	}
	return false
}

// A warning per workspace where the group doesn't tag whoever is on call, and per workspace naming who it leaves out.
func slackGroupWarnings(handle string, groups *[]onCallSlackGroup) diag.Diagnostics {
	if groups == nil {
		return nil
	}
	var warnings diag.Diagnostics
	path := cty.GetAttrPath("slack_groups")
	for _, group := range *groups {
		workspace := stringValue(group.SlackWorkspaceID)
		switch status := stringValue(group.Status); status {
		case "active":
		case "pending":
			warnings = append(warnings, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Slack user group @%s is still being set up in workspace %s", handle, workspace),
				Detail:        fmt.Sprintf("Better Stack hadn't finished setting it up after %s. The next refresh shows how Slack took it in slack_groups.", slackGroupsSetupWait),
				AttributePath: path,
			})
		default:
			detail := stringValue(group.Error)
			if detail == "" {
				detail = fmt.Sprintf("Its status is %s.", status)
			}
			warnings = append(warnings, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Slack user group @%s doesn't work in workspace %s", handle, workspace),
				Detail:        detail,
				AttributePath: path,
			})
		}
		if len(group.UnconnectedUsers) == 0 {
			continue
		}
		names := make([]string, 0, len(group.UnconnectedUsers))
		for _, user := range group.UnconnectedUsers {
			name := strings.TrimSpace(stringValue(user.FirstName) + " " + stringValue(user.LastName))
			names = append(names, fmt.Sprintf("%s (%s)", name, stringValue(user.Email)))
		}
		summary := fmt.Sprintf("A responder won't be tagged by @%s in Slack workspace %s", handle, workspace)
		if len(names) > 1 {
			summary = fmt.Sprintf("Some responders won't be tagged by @%s in Slack workspace %s", handle, workspace)
		}
		warnings = append(warnings, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       summary,
			Detail:        fmt.Sprintf("They can connect by using the same e-mail in Slack and Better Stack, or in their account settings. Not connected yet: %s.", strings.Join(names, ", ")),
			AttributePath: path,
		})
	}
	return warnings
}

func resourceOnCallCalendarUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in onCallCalendar
	for _, e := range onCallCalendarRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	// Neither load nor the planned value carries a configured "", which stops tagging.
	if old, _ := d.GetChange("slack_group_handle"); old.(string) != "" && slackGroupHandleConfiguredEmpty(d.GetRawConfig()) {
		cleared := ""
		in.SlackGroupHandle = &cleared
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

	var rotation *onCallRotation
	if inRotations, ok := d.GetOk("on_call_rotation"); ok && len(inRotations.([]interface{})) > 0 && d.HasChange("on_call_rotation") {
		var outRotation onCallRotation
		if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s/rotation", url.PathEscape(d.Id())), onCallRotationRequest(inRotations.([]interface{})[0]), &outRotation); err != nil {
			return err
		}
		rotation = &outRotation
	} else {
		var derr diag.Diagnostics
		if rotation, derr = readOnCallRotation(ctx, meta, d.Id()); derr != nil {
			return derr
		}
	}

	return onCallCalendarApplied(ctx, d, meta, &out.Data.Attributes, out.Data.Relationships, out.Included, rotation)
}

func resourceOnCallCalendarDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/on-calls/%s", url.PathEscape(d.Id())))
}
