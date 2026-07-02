# Minimal HTTP monitor - checks that the URL returns a 2XX status code
resource "betteruptime_monitor" "simple" {
  url          = "https://example.com"
  monitor_type = "status"
}

# Monitor that passes only when the endpoint returns a specific set of status codes
resource "betteruptime_monitor" "expected_status_code" {
  url                   = "https://example.com"
  monitor_type          = "expected_status_code"
  monitor_group_id      = betteruptime_monitor_group.this.id
  expected_status_codes = [200, 201, 204]
}

# Keyword monitor - alerts when the expected text disappears from the page
resource "betteruptime_monitor" "keyword" {
  url              = "https://example.com"
  monitor_type     = "keyword"
  monitor_group_id = betteruptime_monitor_group.this.id
  required_keyword = "Better Stack"
}

# Keyword absence monitor - the inverse of keyword, alerts when the unwanted text appears on the page
resource "betteruptime_monitor" "keyword_absence" {
  url              = "https://example.com"
  monitor_type     = "keyword_absence"
  monitor_group_id = betteruptime_monitor_group.this.id
  required_keyword = "Error 500"
}

# Ping monitor - checks that the host is reachable via ICMP ping
resource "betteruptime_monitor" "ping" {
  url              = "example.com"
  monitor_type     = "ping"
  monitor_group_id = betteruptime_monitor_group.this.id
}

# TCP monitor - checks that a port is accepting connections
resource "betteruptime_monitor" "tcp" {
  url              = "example.com"
  monitor_type     = "tcp"
  monitor_group_id = betteruptime_monitor_group.this.id
  port             = "443"
}

# DNS monitor - checks that the DNS server at url answers queries for the domain in request_body
resource "betteruptime_monitor" "dns" {
  url              = "1.1.1.1"
  monitor_type     = "dns"
  request_body     = "example.com"
  monitor_group_id = betteruptime_monitor_group.this.id
}

# Playwright browser monitor running a scripted scenario
resource "betteruptime_monitor" "playwright" {
  scenario_name     = "Better Stack Homepage"
  monitor_type      = "playwright"
  monitor_group_id  = betteruptime_monitor_group.this.id
  playwright_script = <<-EOT
    import { test, expect } from '@playwright/test';

    test('has title and contact email', async ({ page }) => {
      await page.goto('https://betterstack.com/')
      await expect(page).toHaveTitle(/Better Stack/)
      await page.goto('https://betterstack.com/contact')
      await expect(page.locator('body')).toContainText(process.env.EMAIL)
    });
  EOT
  request_timeout   = 60
  environment_variables = {
    EMAIL = "hello@betterstack.com"
  }
}

# Comprehensive HTTP "status" monitor - proxy, custom header, expiry alerts and a maintenance window
resource "betteruptime_monitor" "status" {
  # Own registrable domain - expiration_policy_id is a per-domain setting (subdomains
  # share it too), so a dedicated domain avoids clashing with the example.com monitors - use your own domain
  url                  = "https://my-${random_pet.unique.id}-domain.com"
  monitor_type         = "status"
  monitor_group_id     = betteruptime_monitor_group.this.id
  expiration_policy_id = betteruptime_policy.this.id

  # Warn 30 days before the domain registration expires (-1 disables the check)
  domain_expiration = 30

  # Warn 30 days before the SSL certificate expires (-1 disables the check)
  ssl_expiration = 30

  proxy_host = "user:pass@proxy.example.com"
  proxy_port = 8080

  # Check every 30 seconds
  check_frequency = 30

  # Confirm the failure for 60 seconds before opening an incident
  confirmation_period = 60

  # Require 3 minutes of recovery before auto-resolving
  recovery_period = 180

  # Send a POST instead of the default GET
  http_method = "POST"

  # JSON body sent with the request
  request_body = jsonencode({ ping = true })

  # Treat 3xx responses as down instead of following them
  follow_redirects = false

  # Probe from the US and EU
  regions = ["us", "eu"]

  # Force IPv4 probing
  ip_version = "ipv4"

  # Fail the check if there is no response within 30s
  request_timeout = 30

  # Spoken name used on phone-call alerts
  pronounceable_name = "Example API"

  # Escalate downtime through this policy
  policy_id = betteruptime_policy.this.id

  # Wait 3 minutes before escalating to the whole team
  team_wait = 180

  # Nightly maintenance window - checks are suppressed inside it
  maintenance_from = "01:00:00"
  maintenance_to   = "03:00:00"
  maintenance_days = ["sat", "sun"]
  # Rails timezone name, as the API stores it
  maintenance_timezone = "Berlin"

  request_headers = [
    {
      "name" : "X-Source",
      "value" : "terraform"
    },
  ]
}
