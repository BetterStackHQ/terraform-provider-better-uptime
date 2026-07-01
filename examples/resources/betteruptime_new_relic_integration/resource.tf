# Better Stack receives alerts from New Relic through a generated webhook URL
resource "betteruptime_new_relic_integration" "this" {
  name           = "Terraform New Relic Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  policy_id      = betteruptime_policy.this.id # Route alerts through this escalation policy
  alerting_rule  = "alert_and_warn"            # Open incidents for both alarms and warnings
}
