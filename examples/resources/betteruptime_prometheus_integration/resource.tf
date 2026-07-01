# Better Stack receives alerts from Prometheus through a generated webhook URL.
resource "betteruptime_prometheus_integration" "this" {
  name           = "Terraform Prometheus Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
}
