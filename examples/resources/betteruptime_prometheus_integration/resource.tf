# Better Stack receives alerts from Prometheus through a generated webhook URL

resource "betteruptime_prometheus_integration" "this" {
  name           = "Terraform Prometheus Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}

resource "betteruptime_prometheus_integration" "with_policy" {
  name      = "Terraform Prometheus Integration with custom policy"
  policy_id = betteruptime_policy.this.id
}

# Point Prometheus Alertmanager at this URL to deliver alerts to Better Stack

output "prometheus_integration_webhook_url" {
  value = betteruptime_prometheus_integration.this.webhook_url
}
