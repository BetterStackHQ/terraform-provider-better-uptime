# Look up an existing incoming webhook by name
data "betteruptime_incoming_webhook" "existing" {
  name = "My Existing Incoming Webhook"
}

# The endpoint URL where Better Stack receives this webhook
# Named distinctly from the resource example's incoming_webhook_url so both
# coexist in the combined E2E config
output "existing_incoming_webhook_url" {
  value = data.betteruptime_incoming_webhook.existing.url
}
