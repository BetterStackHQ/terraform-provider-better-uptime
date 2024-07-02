provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

resource "random_id" "status_page_subdomain" {
  byte_length = 8
  prefix      = "tf-status-"
}

resource "betteruptime_status_page" "this" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"
  subdomain    = coalesce(var.betteruptime_status_page_subdomain, random_id.status_page_subdomain.hex)
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
