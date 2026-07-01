# Comprehensive HTTP "status" monitor - proxy, custom header, expiry alerts and a maintenance window
resource "betteruptime_monitor" "status" {
  # Own registrable domain - expiration_policy_id is a per-domain setting (subdomains
  # share it too), so a dedicated domain avoids clashing with the example.com monitors
  url                  = "https://${random_pet.unique.id}.com"
  monitor_type         = "status"
  monitor_group_id     = betteruptime_monitor_group.this.id
  expiration_policy_id = betteruptime_policy.this.id
  domain_expiration    = 30 # Warn 30 days before the domain registration expires
  ssl_expiration       = 30 # Warn 30 days before the SSL certificate expires
  proxy_host           = "user:pass@proxy.example.com"
  proxy_port           = 8080
  check_frequency      = 30                          # Check every 30 seconds
  recovery_period      = 180                         # Require 3 minutes of recovery before auto-resolving
  http_method          = "POST"                      # Send a POST instead of the default GET
  request_body         = jsonencode({ ping = true }) # JSON body sent with the request
  regions              = ["us", "eu"]                # Probe from the US and EU
  ip_version           = "ipv4"                      # Force IPv4 probing
  request_timeout      = 30                          # Fail the check if there is no response within 30s
  pronounceable_name   = "Example API"               # Spoken name used on phone-call alerts
  policy_id            = betteruptime_policy.this.id # Escalate downtime through this policy
  team_wait            = 180                         # Wait 3 minutes before escalating to the whole team
  maintenance_from     = "01:00:00"                  # Nightly maintenance window - checks are suppressed inside it
  maintenance_to       = "03:00:00"
  maintenance_days     = ["sat", "sun"]
  maintenance_timezone = "Berlin" # Rails timezone name, as the API stores it
  request_headers = [
    {
      "name" : "X-Source",
      "value" : "terraform"
    },
  ]
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

# TCP monitor - checks that a port is accepting connections
resource "betteruptime_monitor" "tcp" {
  url              = "example.com"
  monitor_type     = "tcp"
  monitor_group_id = betteruptime_monitor_group.this.id
  port             = "443"
}

# DNS monitor - checks that a record resolves to the expected value
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

    test('has title', async ({ page }) => {
      await page.goto('https://betterstack.com/')
      await expect(page).toHaveTitle(/Better Stack/)
    });
  EOT
  request_timeout   = 60
  environment_variables = {
    EMAIL = "hello@betterstack.com"
  }
}
