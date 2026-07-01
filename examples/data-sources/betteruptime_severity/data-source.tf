# Look up a severity (urgency) by name.
data "betteruptime_severity" "existing" {
  name       = betteruptime_severity.this.name
  depends_on = [betteruptime_severity.this]
}
