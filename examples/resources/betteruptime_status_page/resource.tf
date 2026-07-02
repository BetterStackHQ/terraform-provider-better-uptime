# Minimal status page

resource "betteruptime_status_page" "simple" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"

  # Random suffix keeps the subdomain globally unique - pick your own, e.g. "acme-status"
  subdomain = "min-${random_id.status_page_subdomain.hex}"
}

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

  # Use the modern status-page design
  design = "v2"

  # Light theme (applies with design v2)
  theme = "light"

  # Vertical layout (applies with design v2)
  layout = "vertical"

  # Show 90 days of history
  history = 90

  # Hide incidents shorter than 5 minutes
  min_incident_length = 300

  # Keep the page out of search engines
  hide_from_search_engines = true

  # Auto-generate downtime reports
  automatic_reports = true

  # Banner announcement text
  announcement = "All systems operational."

  # Show the announcement in the JS embed
  announcement_embed_visible = true

  # Link target for the embedded announcement
  announcement_embed_link = "https://example.com/status"

  # Style the announcement embed
  announcement_embed_css = ".announcement { font-weight: bold; }"

  # Custom branding CSS
  custom_css = ".page { --brand: #0f172a; }"

  # Google Analytics measurement ID
  google_analytics_id = "G-XXXXXXXXXX"

  # Remove the "Powered by Better Stack" footer - available on higher plans
  whitelabeled = true

  # TODO: point at refs/heads/master instead of the branch before merging
  # Company logo
  logo_url = "https://raw.githubusercontent.com/BetterStackHQ/terraform-provider-better-uptime/refs/heads/petr/uptime-examples-docs-e2e/examples/resources/betteruptime_status_page/logo_black_text.png"

  # Logo for the dark theme
  dark_logo_url = "https://raw.githubusercontent.com/BetterStackHQ/terraform-provider-better-uptime/refs/heads/petr/uptime-examples-docs-e2e/examples/resources/betteruptime_status_page/logo_white_text.png"

  # Place the page in a status-page group
  status_page_group_id = betteruptime_status_page_group.this.id

  # Password-protect the page
  password_enabled = true

  # Required when password_enabled is true
  password = "s3cret"

  # Custom navigation links (design v2)
  navigation_links {
    text = "Home"
    href = "https://example.com"
  }
  navigation_links {
    text = "Incidents"
    href = "/incidents"
  }
}
