# Better Stack receives alerts from Azure through a generated webhook URL.
resource "betteruptime_azure_integration" "this" {
  name           = "Terraform Azure Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}
