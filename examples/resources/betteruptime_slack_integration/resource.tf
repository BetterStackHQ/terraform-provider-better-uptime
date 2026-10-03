# Slack integrations are connected in the Better Stack web UI (it's a Slack sign-in),
# so Terraform configures an existing one: look it up by its channel, then adopt it
# by its better_stack_id. Destroying the resource leaves the integration connected.
data "betteruptime_slack_integration" "incidents_channel" {
  slack_channel_name = "#my-existing-slack-channel"
}

# The escalation policy the channel can run, and a monitor whose status page updates it gets
data "betteruptime_policy" "slack_alerts" {
  name = "My Existing Escalation Policy"
}
data "betteruptime_monitor" "slack_status_updates" {
  url = "https://betterstack.com"
}

resource "betteruptime_slack_integration" "this" {
  better_stack_id = data.betteruptime_slack_integration.incidents_channel.id

  # One message per incident; thread and channel need a paid plan
  integration_type = "verbose"

  # Offer only the policy below, instead of every team and policy, when alerting from Slack
  alert_all_teams_and_policies = false
  alert_option {
    type = "policy"
    id   = data.betteruptime_policy.slack_alerts.id
  }

  # Post status page updates of the monitor to the channel
  post_status_page_updates = true
  status_page_update_resource {
    type = "monitor"
    id   = data.betteruptime_monitor.slack_status_updates.id
  }

  on_call_notifications                 = true
  notify_on_pause                       = false
  notify_on_resolve                     = true
  post_timeline_events                  = true
  post_incident_metadata                = true
  direct_message_on_mention             = false
  ai_sre_on_mention                     = false
  invite_new_slack_users                = false
  restrict_incident_creation_to_channel = false
  notify_alongside_primary_responder    = false
}
