# Better Stack receives alerts from Azure through a generated webhook URL
resource "betteruptime_azure_integration" "this" {
  name           = "Terraform Azure Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}

# Point Azure at this URL to deliver alerts to Better Stack
output "azure_integration_webhook_url" {
  value = betteruptime_azure_integration.this.webhook_url
}
