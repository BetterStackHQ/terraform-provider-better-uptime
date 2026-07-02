# Better Stack receives alerts from Elastic through a generated webhook URL
resource "betteruptime_elastic_integration" "this" {
  name           = "Terraform Elastic Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = true                        # Bypass Do not Disturb on the mobile app
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}

# Point Elastic at this URL to deliver alerts to Better Stack
output "elastic_integration_webhook_url" {
  value = betteruptime_elastic_integration.this.webhook_url
}
