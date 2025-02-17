output "betteruptime_status_page_url" {
  value = "https://${betteruptime_status_page.this.subdomain}.betteruptime.com"
}

output "escalation_policy_id" {
  value = betteruptime_policy.this.id
}

output "betteruptime_email_integration_address" { value = betteruptime_email_integration.this.email_address }

output "betteruptime_incoming_webhook_url" { value = betteruptime_incoming_webhook.this.url }

output "betteruptime_elastic_integration_webhook_url" { value = betteruptime_elastic_integration.this.webhook_url }

output "betteruptime_datadog_integration_webhook_url" { value = betteruptime_datadog_integration.this.webhook_url }
