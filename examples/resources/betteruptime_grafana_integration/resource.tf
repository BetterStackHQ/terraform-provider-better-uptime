# Better Stack receives alerts from Grafana through a generated webhook URL.
resource "betteruptime_grafana_integration" "this" {
  name           = "Terraform Grafana Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}
