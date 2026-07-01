# Look up an incoming webhook by name.
data "betteruptime_incoming_webhook" "existing" {
  name       = betteruptime_incoming_webhook.this.name
  depends_on = [betteruptime_incoming_webhook.this]
}
