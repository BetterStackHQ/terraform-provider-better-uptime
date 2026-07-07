# Better Stack receives alerts from Datadog through a generated webhook URL
resource "betteruptime_datadog_integration" "this" {
  name           = "Terraform Datadog Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false

  # Open incidents for both alerts and warnings
  alerting_rule = "alert_and_warn"
}

resource "betteruptime_datadog_integration" "with_policy" {
  name      = "Terraform Datadog Integration with custom policy"
  policy_id = betteruptime_policy.this.id
}

# Point Datadog at this URL to deliver alerts to Better Stack
output "datadog_integration_webhook_url" {
  value = betteruptime_datadog_integration.this.webhook_url
}
