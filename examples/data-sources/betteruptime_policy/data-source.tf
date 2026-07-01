# Look up an escalation policy by name
data "betteruptime_policy" "existing" {
  name       = betteruptime_policy.this.name
  depends_on = [betteruptime_policy.this]
}
