# Better Stack receives alerts from Google Monitoring through a generated webhook URL
resource "betteruptime_google_monitoring_integration" "this" {
  name           = "Terraform Google Monitoring Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}

resource "betteruptime_google_monitoring_integration" "with_policy" {
  name      = "Terraform Google Monitoring Integration with custom policy"
  policy_id = betteruptime_policy.this.id
}

# Point Google Monitoring at this URL to deliver alerts to Better Stack
output "google_monitoring_integration_webhook_url" {
  value = betteruptime_google_monitoring_integration.this.webhook_url
}
