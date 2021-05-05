provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

resource "betteruptime_status_page" "this" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"
  subdomain    = var.betteruptime_status_page_subdomain
}

resource "betteruptime_monitor" "this" {
  url          = "https://example.com"
  monitor_type = "status"
}

resource "betteruptime_status_page_resource" "monitor" {
  status_page_id = betteruptime_status_page.this.id
  resource_id    = betteruptime_monitor.this.id
  resource_type  = "Monitor"
  public_name    = "example.com site"
}
