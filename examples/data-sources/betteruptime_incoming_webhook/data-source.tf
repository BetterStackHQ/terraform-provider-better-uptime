# Look up an existing incoming webhook by name
data "betteruptime_incoming_webhook" "existing" {
  name = "My Existing Incoming Webhook"
}

# The webhook ingest URL stays out of CI logs - read it as data...existing.url
output "existing_incoming_webhook_id" {
  value = data.betteruptime_incoming_webhook.existing.id
}
