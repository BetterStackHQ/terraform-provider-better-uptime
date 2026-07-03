# Better Stack receives alerts from New Relic through a generated webhook URL
resource "betteruptime_new_relic_integration" "this" {
  name           = "Terraform New Relic Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false

  # Open incidents for both alerts and warnings
  alerting_rule = "alert_and_warn"
}

resource "betteruptime_new_relic_integration" "with_policy" {
  name      = "Terraform New Relic Integration with custom policy"
  policy_id = betteruptime_policy.this.id
}

# Point New Relic at this URL to deliver alerts to Better Stack
output "new_relic_integration_webhook_url" {
  value = betteruptime_new_relic_integration.this.webhook_url
}
