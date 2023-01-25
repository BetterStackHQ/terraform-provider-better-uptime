output "betteruptime_status_page_url" {
  value = "https://${betteruptime_status_page.this.subdomain}.betteruptime.com"
}

output "escalation_policy_id" {
  value = betteruptime_policy.this.id
}

output "betteruptime_email_integration_address" { value = betteruptime_email_integration.this.email_address }

output "betteruptime_incoming_webhook_url" { value = betteruptime_incoming_webhook.this.url }
