# Better Stack receives alerts from Google Monitoring through a generated webhook URL
resource "betteruptime_google_monitoring_integration" "this" {
  name           = "Terraform Google Monitoring Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}

# Point Google Monitoring at this URL to deliver alerts to Better Stack
output "google_monitoring_integration_webhook_url" {
  value = betteruptime_google_monitoring_integration.this.webhook_url
}
