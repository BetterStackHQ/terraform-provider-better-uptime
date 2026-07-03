# Minimal status page
resource "betteruptime_status_page" "simple" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"

  # Random suffix keeps the subdomain globally unique - pick your own, e.g. "acme-status"
  subdomain = "tf-status-simple-${random_id.status_page_subdomain.hex}"
}

# Status page with the commonly used options
resource "betteruptime_status_page" "this" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  contact_url  = "mailto:support@example.com"
  timezone     = "Eastern Time (US & Canada)"
  subdomain    = "tf-status-${random_id.status_page_subdomain.hex}"
  subscribable = true

  # Set to your own domain, e.g. "status.example.com" - setting it back to "" removes the custom domain again
  custom_domain = ""

  # Place the page in a status-page group
  status_page_group_id = betteruptime_status_page_group.this.id

  # Show 90 days of history, hiding incidents shorter than 5 minutes
  history             = 90
  min_incident_length = 300

  # Auto-generate downtime reports and keep the page out of search engines
  automatic_reports        = true
  hide_from_search_engines = true
}

# Styled status page - branding, announcement and custom navigation
resource "betteruptime_status_page" "styled" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"
  subdomain    = "tf-status-styled-${random_id.status_page_subdomain.hex}"

  # The modern design with a light theme and vertical layout
  design = "v2"
  theme  = "light"
  layout = "vertical"

  # Company logos for the light and dark theme
  logo_url      = "https://raw.githubusercontent.com/BetterStackHQ/terraform-provider-better-uptime/refs/heads/master/examples/resources/betteruptime_status_page/logo_black_text.png"
  dark_logo_url = "https://raw.githubusercontent.com/BetterStackHQ/terraform-provider-better-uptime/refs/heads/master/examples/resources/betteruptime_status_page/logo_white_text.png"

  # Banner announcement, also shown in the JavaScript embed
  announcement               = "All systems operational."
  announcement_embed_visible = true
  announcement_embed_link    = "https://example.com/status"
  announcement_embed_css     = ".announcement { font-weight: bold; }"

  custom_css          = ".page { --brand: #0f172a; }"
  google_analytics_id = "G-XXXXXXXXXX"

  # Remove the "Powered by Better Stack" footer - available on higher plans
  whitelabeled = true

  # Custom navigation links (custom_domain and custom_javascript are also
  # available once a CNAME points at Better Stack)
  navigation_links {
    text = "Home"
    href = "https://example.com"
  }
  navigation_links {
    text = "Incidents"
    href = "/incidents"
  }
}

# Restricted status page - visible only from allowed networks, with a password
resource "betteruptime_status_page" "secure" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"
  subdomain    = "tf-status-secure-${random_id.status_page_subdomain.hex}"

  # Comment lines are supported in the allowlist
  ip_allowlist = [
    "# Office network",
    "192.168.1.0/24",
    "2001:0db8:85a3::/64",
  ]

  # Password-protect the page
  password_enabled = true
  password         = "s3cret"
}
