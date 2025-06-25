package provider

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
		Description: "Type of the Slack integration. Possible values: legacy, verbose, thread, channel",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"on_call_notifications": {
		Description: "Whether to post a notification when the current on-call person changes.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
}

type slackIntegration struct {
	Id                  *string `json:"id,omitempty"`
	SlackTeamId         *string `json:"slack_team_id,omitempty"`
	SlackTeamName       *string `json:"slack_team_name,omitempty"`
	SlackChannelId      *string `json:"slack_channel_id,omitempty"`
	SlackChannelName    *string `json:"slack_channel_name,omitempty"`
	SlackStatus         *string `json:"slack_status,omitempty"`
	IntegrationTyp      *string `json:"integration_type,omitempty"`
	OnCallNotifications *bool   `json:"on_call_notifications,omitempty"`
	TeamName            *string `json:"team_name,omitempty"`
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
		{k: "on_call_notifications", v: &in.OnCallNotifications},
	}
}

func slackIntegrationCopyAttrs(d *schema.ResourceData, in *slackIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range slackIntegrationRef(in) {
		value := reflect.Indirect(reflect.ValueOf(e.v)).Interface()
		if err := d.Set(e.k, value); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}
