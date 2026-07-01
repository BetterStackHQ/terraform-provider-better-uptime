resource "betteruptime_elastic_integration" "this" {
  name           = "Terraform Elastic Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = true
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}
