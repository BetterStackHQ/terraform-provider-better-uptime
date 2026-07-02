# Metadata on a monitor
resource "betteruptime_metadata" "monitor_playwright" {
  owner_id   = betteruptime_monitor.playwright.id
  owner_type = "Monitor"
  key        = "E-mail"
  metadata_value {
    value = "hello@betterstack.com"
  }
}

# A User-typed metadata value, referenced by policy metadata branching
resource "betteruptime_metadata" "assigned_user" {
  owner_type = "EmailIntegration"
  owner_id   = betteruptime_email_integration.this.id
  key        = "Assigned User"
  metadata_value {
    type  = "User"
    email = "petr@betterstack.com"
  }
}

# A Policy-typed metadata value (points at an escalation policy by id)
resource "betteruptime_metadata" "assigned_policy" {
  owner_type = "EmailIntegration"
  owner_id   = betteruptime_email_integration.this.id
  key        = "Assigned Policy"
  metadata_value {
    type    = "Policy"
    item_id = betteruptime_policy.this.id
  }
}

# Metadata on a Heartbeat owner, with a Policy referenced by name instead of by id
resource "betteruptime_metadata" "runbook_owner" {
  owner_type = "Heartbeat"
  owner_id   = betteruptime_heartbeat.this.id
  key        = "runbook-owner"
  metadata_value {
    type = "Policy"
    name = "My Existing Escalation Policy" # Reference a policy by name instead of by id
  }
}
