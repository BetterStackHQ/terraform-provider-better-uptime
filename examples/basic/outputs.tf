output "betteruptime_status_page_url" {
  value = "https://${betteruptime_status_page.this.subdomain}.betteruptime.com"
}

output "betteruptime_email_integration_address" { value = betteruptime_email_integration.this.email_address }
