output "betteruptime_status_page_url" {
  value = "https://${betteruptime_status_page.this.subdomain}.betteruptime.com"
}

output "escalation_policy_id" {
  value = betteruptime_policy.this.id
}
