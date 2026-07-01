# Look up a connected Slack integration by channel. Requires Slack to be connected
# in Better Stack, so this example is documented but excluded from the combined E2E.
data "betteruptime_slack_integration" "existing" {
  slack_channel_name = "alerts"
}
