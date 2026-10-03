package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// The list attributes are sets: the API keeps the rows it already has when a list is replaced, so
// it may return them in a different order than they were sent.
var slackIntegrationChannelInviteeSchema = map[string]*schema.Schema{
	"type": {
		Description:  "Type of the invitee. Possible values: user, on_call_calendar, slack_user_group",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"user", "on_call_calendar", "slack_user_group"}, false),
	},
	"id": {
		Description: "ID of the user, on-call calendar or Slack user group.",
		Type:        schema.TypeInt,
		Required:    true,
	},
}

var slackIntegrationAlertOptionSchema = map[string]*schema.Schema{
	"type": {
		Description:  "Whom the option alerts. Possible values: team (alerts the team's members on the channels below), policy (runs the escalation policy)",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"team", "policy"}, false),
	},
	"id": {
		Description: "ID of the team or escalation policy.",
		Type:        schema.TypeInt,
		Required:    true,
	},
	"call": {
		Description: "Whether to alert the team by phone call. Only for type team.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"sms": {
		Description: "Whether to alert the team by SMS. Only for type team.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"email": {
		Description: "Whether to alert the team by e-mail. Only for type team.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"push": {
		Description: "Whether to alert the team by push notification. Only for type team.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"critical_alert": {
		Description: "Whether to alert the team by critical push notification. Only for type team.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
}

var slackIntegrationStatusPageUpdateResourceSchema = map[string]*schema.Schema{
	"type": {
		Description:  "Type of the resource. Possible values: status_page, monitor, heartbeat, monitor_group, heartbeat_group",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"status_page", "monitor", "heartbeat", "monitor_group", "heartbeat_group"}, false),
	},
	"id": {
		Description: "ID of the resource.",
		Type:        schema.TypeInt,
		Required:    true,
	},
}

var slackIntegrationSchema = map[string]*schema.Schema{
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
		Description: "The ID of this Slack integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"slack_team_id": {
		Description: "Slack ID of the connected team.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"slack_team_name": {
		Description: "Name of the connected Slack team.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"slack_channel_id": {
		Description: "Slack ID of the connected channel.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"slack_channel_name": {
		Description: "Name of the connected Slack channel.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"slack_status": {
		Description: "Status of the connected Slack account. Possible values: active, account_inactive",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"integration_type": {
		Description:      "Type of the Slack integration. Possible values: legacy, verbose, thread, channel. Only verbose, thread and channel can be set; thread and channel need a paid plan, and legacy is what integrations connected before those types existed still hold.",
		Type:             schema.TypeString,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"verbose", "thread", "channel"}, false)),
	},
	"channel_invitee": {
		Description: "Who to invite to the incident channels a channel-type integration creates. Removing every block clears the list.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Resource{Schema: slackIntegrationChannelInviteeSchema},
	},
	"private_channel": {
		Description: "Whether a channel-type integration creates private incident channels. Private channels need at least one channel_invitee.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"channel_display_name": {
		Description: "Name used for the incident channels a channel-type integration creates. Removing it clears the name.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"alert_all_teams_and_policies": {
		Description: "Whether anyone in the channel may alert any team or run any escalation policy. When false, only the alert_option entries are offered.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"alert_option": {
		Description: "Teams and escalation policies the channel can alert when alert_all_teams_and_policies is false. Removing every block clears the list.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Resource{Schema: slackIntegrationAlertOptionSchema},
	},
	"invite_new_slack_users": {
		Description: "Whether to invite new Slack users into your Better Stack team as responders. Invited responders may be billed.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"on_call_notifications": {
		Description: "Whether to post a notification when the current on-call person changes.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"notify_on_pause": {
		Description: "Whether to post a notification when a resource is paused.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"notify_on_resolve": {
		Description: "Whether to post a notification when an incident is resolved.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"direct_message_on_mention": {
		Description: "Whether to direct message Slack users when they are tagged in incident comments.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"ai_sre_on_mention": {
		Description: "Whether to start an AI SRE conversation when Better Stack is tagged in Slack.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"restrict_incident_creation_to_channel": {
		Description: "Whether incidents can be created from Slack only in this channel (or, for a channel-type integration, only by its members).",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"post_timeline_events": {
		Description: "Whether to post incident timeline events about e-mails, phone calls, and integrations.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"post_incident_metadata": {
		Description: "Whether to post the incident's metadata.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"post_status_page_updates": {
		Description: "Whether to post status page updates of the status_page_update_resource entries. Turning it off clears them.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"status_page_update_resource": {
		Description: "Status pages and resources whose status page updates are posted, when post_status_page_updates is true. Removing every block clears the list.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Resource{Schema: slackIntegrationStatusPageUpdateResourceSchema},
	},
	"notify_alongside_primary_responder": {
		Description: "Whether this integration should be notified alongside the primary responder when no escalation policy is configured.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
}

func newSlackIntegrationResource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range slackIntegrationSchema {
		s[k] = v
	}
	// The integration already exists in its team, so there's nothing to pick a team for.
	delete(s, "team_name")
	s["better_stack_id"] = &schema.Schema{
		Description: "Due to required authentication in Slack, the integration has to be connected and removed in Better Stack web UI. Set the ID of the Slack integration to manage, and it will be adopted during resource creation. Destroying the resource only removes it from the Terraform state.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
	}
	return &schema.Resource{
		CreateContext: slackIntegrationCreate,
		ReadContext:   slackIntegrationRead,
		UpdateContext: slackIntegrationUpdate,
		DeleteContext: slackIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Configures a Slack integration connected in the Better Stack web UI. https://betterstack.com/docs/uptime/api/slack-integrations/",
		Schema:      s,
	}
}

type slackIntegrationReference struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type slackIntegrationAlertOption struct {
	Type          string `json:"type"`
	ID            int    `json:"id"`
	Call          *bool  `json:"call,omitempty"`
	SMS           *bool  `json:"sms,omitempty"`
	Email         *bool  `json:"email,omitempty"`
	Push          *bool  `json:"push,omitempty"`
	CriticalAlert *bool  `json:"critical_alert,omitempty"`
}

type slackIntegration struct {
	Id                                *string                        `json:"id,omitempty"`
	SlackTeamId                       *string                        `json:"slack_team_id,omitempty"`
	SlackTeamName                     *string                        `json:"slack_team_name,omitempty"`
	SlackChannelId                    *string                        `json:"slack_channel_id,omitempty"`
	SlackChannelName                  *string                        `json:"slack_channel_name,omitempty"`
	SlackStatus                       *string                        `json:"slack_status,omitempty"`
	IntegrationTyp                    *string                        `json:"integration_type,omitempty"`
	ChannelInvitees                   *[]slackIntegrationReference   `json:"channel_invitees,omitempty"`
	PrivateChannel                    *bool                          `json:"private_channel,omitempty"`
	ChannelDisplayName                *string                        `json:"channel_display_name,omitempty"`
	AlertAllTeamsAndPolicies          *bool                          `json:"alert_all_teams_and_policies,omitempty"`
	AlertOptions                      *[]slackIntegrationAlertOption `json:"alert_options,omitempty"`
	InviteNewSlackUsers               *bool                          `json:"invite_new_slack_users,omitempty"`
	OnCallNotifications               *bool                          `json:"on_call_notifications,omitempty"`
	NotifyOnPause                     *bool                          `json:"notify_on_pause,omitempty"`
	NotifyOnResolve                   *bool                          `json:"notify_on_resolve,omitempty"`
	DirectMessageOnMention            *bool                          `json:"direct_message_on_mention,omitempty"`
	AiSreOnMention                    *bool                          `json:"ai_sre_on_mention,omitempty"`
	RestrictIncidentCreationToChannel *bool                          `json:"restrict_incident_creation_to_channel,omitempty"`
	PostTimelineEvents                *bool                          `json:"post_timeline_events,omitempty"`
	PostIncidentMetadata              *bool                          `json:"post_incident_metadata,omitempty"`
	PostStatusPageUpdates             *bool                          `json:"post_status_page_updates,omitempty"`
	StatusPageUpdateResources         *[]slackIntegrationReference   `json:"status_page_update_resources,omitempty"`
	NotifyAlongsidePrimaryResponder   *bool                          `json:"notify_alongside_primary_responder,omitempty"`
	TeamName                          *string                        `json:"team_name,omitempty"`
}

type slackIntegrationHTTPResponse struct {
	Data struct {
		ID         string           `json:"id"`
		Attributes slackIntegration `json:"attributes"`
	} `json:"data"`
}

func slackIntegrationRef(in *slackIntegration) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "slack_team_id", v: &in.SlackTeamId},
		{k: "slack_team_name", v: &in.SlackTeamName},
		{k: "slack_channel_id", v: &in.SlackChannelId},
		{k: "slack_channel_name", v: &in.SlackChannelName},
		{k: "slack_status", v: &in.SlackStatus},
		{k: "integration_type", v: &in.IntegrationTyp},
		{k: "private_channel", v: &in.PrivateChannel},
		{k: "channel_display_name", v: &in.ChannelDisplayName},
		{k: "alert_all_teams_and_policies", v: &in.AlertAllTeamsAndPolicies},
		{k: "invite_new_slack_users", v: &in.InviteNewSlackUsers},
		{k: "on_call_notifications", v: &in.OnCallNotifications},
		{k: "notify_on_pause", v: &in.NotifyOnPause},
		{k: "notify_on_resolve", v: &in.NotifyOnResolve},
		{k: "direct_message_on_mention", v: &in.DirectMessageOnMention},
		{k: "ai_sre_on_mention", v: &in.AiSreOnMention},
		{k: "restrict_incident_creation_to_channel", v: &in.RestrictIncidentCreationToChannel},
		{k: "post_timeline_events", v: &in.PostTimelineEvents},
		{k: "post_incident_metadata", v: &in.PostIncidentMetadata},
		{k: "post_status_page_updates", v: &in.PostStatusPageUpdates},
		{k: "notify_alongside_primary_responder", v: &in.NotifyAlongsidePrimaryResponder},
	}
}

// The attributes PATCH accepts; the slack_* ones are read-only.
var slackIntegrationWritableAttributes = map[string]bool{
	"integration_type":                      true,
	"private_channel":                       true,
	"channel_display_name":                  true,
	"alert_all_teams_and_policies":          true,
	"invite_new_slack_users":                true,
	"on_call_notifications":                 true,
	"notify_on_pause":                       true,
	"notify_on_resolve":                     true,
	"direct_message_on_mention":             true,
	"ai_sre_on_mention":                     true,
	"restrict_incident_creation_to_channel": true,
	"post_timeline_events":                  true,
	"post_incident_metadata":                true,
	"post_status_page_updates":              true,
	"notify_alongside_primary_responder":    true,
}

func slackIntegrationCopyAttrs(d *schema.ResourceData, in *slackIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range slackIntegrationRef(in) {
		value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
		if err := d.Set(e.k, value); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	sets := map[string]interface{}{
		"channel_invitee":             slackIntegrationReferencesToSet(in.ChannelInvitees),
		"alert_option":                slackIntegrationAlertOptionsToSet(in.AlertOptions),
		"status_page_update_resource": slackIntegrationReferencesToSet(in.StatusPageUpdateResources),
	}
	for k, v := range sets {
		if err := d.Set(k, v); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func slackIntegrationReferencesToSet(in *[]slackIntegrationReference) []interface{} {
	out := make([]interface{}, 0)
	if in == nil {
		return out
	}
	for _, e := range *in {
		out = append(out, map[string]interface{}{"type": e.Type, "id": e.ID})
	}
	return out
}

func slackIntegrationAlertOptionsToSet(in *[]slackIntegrationAlertOption) []interface{} {
	out := make([]interface{}, 0)
	if in == nil {
		return out
	}
	value := func(b *bool) bool { return b != nil && *b }
	for _, e := range *in {
		out = append(out, map[string]interface{}{
			"type":           e.Type,
			"id":             e.ID,
			"call":           value(e.Call),
			"sms":            value(e.SMS),
			"email":          value(e.Email),
			"push":           value(e.Push),
			"critical_alert": value(e.CriticalAlert),
		})
	}
	return out
}

func slackIntegrationReferencesFromSet(v interface{}) *[]slackIntegrationReference {
	out := make([]slackIntegrationReference, 0)
	for _, e := range v.(*schema.Set).List() {
		m := e.(map[string]interface{})
		out = append(out, slackIntegrationReference{Type: m["type"].(string), ID: m["id"].(int)})
	}
	return &out
}

func slackIntegrationAlertOptionsFromSet(v interface{}) *[]slackIntegrationAlertOption {
	out := make([]slackIntegrationAlertOption, 0)
	for _, e := range v.(*schema.Set).List() {
		m := e.(map[string]interface{})
		option := slackIntegrationAlertOption{Type: m["type"].(string), ID: m["id"].(int)}
		// The API refuses alert channels on an escalation policy, and defaults a team's unsent ones to false.
		if option.Type == "team" {
			flag := func(k string) *bool { b := m[k].(bool); return &b }
			option.Call, option.SMS, option.Email, option.Push, option.CriticalAlert = flag("call"), flag("sms"), flag("email"), flag("push"), flag("critical_alert")
		}
		out = append(out, option)
	}
	return &out
}

func slackIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, ok := d.GetOk("better_stack_id")
	if !ok {
		return diag.Errorf("Due to required authentication in Slack, the integration has to be connected and removed in Better Stack web UI. You can either import the resource, or set the ID of the Slack integration in better_stack_id field and it will be adopted during resource creation.")
	}
	// Creation is not supported for this resource: adopt better_stack_id and apply the whole config.
	d.SetId(id.(string))
	var current slackIntegrationHTTPResponse
	if derr, ok := resourceRead(ctx, meta, slackIntegrationURL(d.Id()), &current); derr != nil {
		return derr
	} else if !ok {
		return diag.Errorf("Slack integration %s not found.", d.Id())
	}
	return slackIntegrationPatch(ctx, d, meta, true)
}

func slackIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out slackIntegrationHTTPResponse
	if derr, ok := resourceRead(ctx, meta, slackIntegrationURL(d.Id()), &out); derr != nil {
		return derr
	} else if !ok {
		d.SetId("") // Disconnected in the web UI.
		return nil
	}
	return slackIntegrationCopyResourceAttrs(d, &out.Data.Attributes)
}

func slackIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return slackIntegrationPatch(ctx, d, meta, false)
}

// Sends the changed attributes, or on adoption every attribute the config sets: a false boolean
// isn't a change from an empty state, but still has to reach the API.
func slackIntegrationPatch(ctx context.Context, d *schema.ResourceData, meta interface{}, adopting bool) diag.Diagnostics {
	config := d.GetRawConfig()
	send := func(k string) bool {
		if adopting {
			// An attribute the API doesn't fill in is cleared when the config leaves it out, so the
			// state converges in one apply.
			return !slackIntegrationSchema[k].Computed || (!config.IsNull() && !config.GetAttr(k).IsNull())
		}
		return d.HasChange(k)
	}

	var in slackIntegration
	for _, e := range slackIntegrationRef(&in) {
		if slackIntegrationWritableAttributes[e.k] && send(e.k) {
			// Not load: it skips a value that was removed, which has to be sent to clear it.
			switch v := e.v.(type) {
			case **string:
				t := d.Get(e.k).(string)
				*v = &t
			case **bool:
				t := d.Get(e.k).(bool)
				*v = &t
			}
		}
	}
	if send("channel_invitee") {
		in.ChannelInvitees = slackIntegrationReferencesFromSet(d.Get("channel_invitee"))
	}
	if send("alert_option") {
		in.AlertOptions = slackIntegrationAlertOptionsFromSet(d.Get("alert_option"))
	}
	if send("status_page_update_resource") {
		in.StatusPageUpdateResources = slackIntegrationReferencesFromSet(d.Get("status_page_update_resource"))
	}

	var out slackIntegrationHTTPResponse
	if derr := resourceUpdate(ctx, meta, slackIntegrationURL(d.Id()), &in, &out); derr != nil {
		return derr
	}
	return slackIntegrationCopyResourceAttrs(d, &out.Data.Attributes)
}

func slackIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Deletion is not supported for this resource: it only leaves the Terraform state.
	d.SetId("")
	return nil
}

func slackIntegrationCopyResourceAttrs(d *schema.ResourceData, in *slackIntegration) diag.Diagnostics {
	derr := slackIntegrationCopyAttrs(d, in)
	// After an import, so a config that sets better_stack_id to the imported ID doesn't replace it.
	if err := d.Set("better_stack_id", d.Id()); err != nil {
		derr = append(derr, diag.FromErr(err)[0])
	}
	return derr
}

func slackIntegrationURL(id string) string {
	return fmt.Sprintf("/api/v2/slack-integrations/%s", url.PathEscape(id))
}
