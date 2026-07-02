# Better Stack receives alerts from Elastic through a generated webhook URL
resource "betteruptime_elastic_integration" "this" {
  name  = "Terraform Elastic Integration"
  call  = false
  sms   = false
  email = true
  push  = true

  # Bypass Do not Disturb on the mobile app
  critical_alert = true
}

resource "betteruptime_elastic_integration" "with_policy" {
  name      = "Terraform Elastic Integration with custom policy"
  policy_id = betteruptime_policy.this.id
}

# Point Elastic at this URL to deliver alerts to Better Stack
output "elastic_integration_webhook_url" {
  value = betteruptime_elastic_integration.this.webhook_url
}
