# Better Stack forwards incidents to PagerDuty using its Events API routing key.
resource "betteruptime_pagerduty_integration" "this" {
  name     = "Terraform PagerDuty Integration"
  key      = "0123456789abcdef0123456789abcdef"
  severity = "critical"

  notify_alongside_primary_responder = false # Don't also notify PagerDuty when no escalation policy is set (default is true)
}
