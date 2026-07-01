# HTTP "status" monitor with a proxy, a custom header and expiration checks tuned off.
resource "betteruptime_monitor" "status" {
  url                  = "https://example.com"
  monitor_type         = "status"
  monitor_group_id     = betteruptime_monitor_group.this.id
  expiration_policy_id = betteruptime_policy.this.id
  domain_expiration    = -1 # Disable domain expiration check
  ssl_expiration       = -1 # Disable SSL expiration check
  proxy_host           = "user:pass@proxy.example.com"
  proxy_port           = 8080
  request_headers = [
    {
      "name" : "X-Source",
      "value" : "terraform"
    },
  ]
}

# DNS monitor - checks that a record resolves to the expected value.
resource "betteruptime_monitor" "dns" {
  url              = "1.1.1.1"
  monitor_type     = "dns"
  request_body     = "example.com"
  monitor_group_id = betteruptime_monitor_group.this.id
}

# Playwright browser monitor running a scripted scenario.
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
