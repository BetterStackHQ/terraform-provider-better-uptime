# Minimal status page - subdomain must be globally unique across Better Stack
resource "betteruptime_status_page" "simple" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"
  subdomain    = "min-${random_id.status_page_subdomain.hex}"
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

  design  = "v2"       # Use the modern status-page design
  theme   = "light"    # Light theme (applies with design v2)
  layout  = "vertical" # Vertical layout (applies with design v2)
  history = 90         # Show 90 days of history

  min_incident_length      = 300  # Hide incidents shorter than 5 minutes
  hide_from_search_engines = true # Keep the page out of search engines
  automatic_reports        = true # Auto-generate downtime reports

  announcement               = "All systems operational."             # Banner announcement text
  announcement_embed_visible = true                                   # Show the announcement in the JS embed
  announcement_embed_link    = "https://example.com/status"           # Link target for the embedded announcement
  announcement_embed_css     = ".announcement { font-weight: bold; }" # Style the announcement embed

  custom_css          = ".page { --brand: #0f172a; }" # Custom branding CSS
  google_analytics_id = "G-XXXXXXXXXX"                # Google Analytics measurement ID

  status_page_group_id = betteruptime_status_page_group.this.id # Place the page in a status-page group

  password_enabled = true     # Password-protect the page
  password         = "s3cret" # Required when password_enabled is true

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
