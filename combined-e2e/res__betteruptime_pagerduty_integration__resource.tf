# Better Stack forwards incidents to PagerDuty using its Events API routing key.
resource "betteruptime_pagerduty_integration" "this" {
  name     = "Terraform PagerDuty Integration"
  key      = "0123456789abcdef0123456789abcdef"
  severity = "critical"
}
