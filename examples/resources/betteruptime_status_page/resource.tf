resource "betteruptime_status_page" "this" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  contact_url  = "mailto:support@example.com"
  timezone     = "Eastern Time (US & Canada)"
  subdomain    = random_id.status_page_subdomain.hex
  subscribable = true
  ip_allowlist = [
    "# Office network",
    "192.168.1.0/24",
    "2001:0db8:85a3::/64",
  ]
}
