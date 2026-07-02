# Look up an existing connected Slack integration by channel name

data "betteruptime_slack_integration" "existing" {
  slack_channel_name = "#my-existing-slack-channel"
}

# The connected Slack team the channel belongs to

output "slack_integration_team_name" {
  value = data.betteruptime_slack_integration.existing.slack_team_name
}
