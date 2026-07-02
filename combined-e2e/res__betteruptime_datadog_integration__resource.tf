resource "betteruptime_datadog_integration" "this" {
  name           = "Terraform Datadog Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
  alerting_rule  = "alert_and_warn"            # Open incidents for both alarms and warnings
}

# Point Datadog at this URL to deliver alerts to Better Stack
output "datadog_integration_webhook_url" {
  value = betteruptime_datadog_integration.this.webhook_url
}
