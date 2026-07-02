# Look up an existing incoming webhook by name
data "betteruptime_incoming_webhook" "existing" {
  name = "My Existing Incoming Webhook"
}

output "existing_incoming_webhook_url" {
  value = data.betteruptime_incoming_webhook.existing.url
}
