# Metadata on a monitor
resource "betteruptime_metadata" "monitor" {
  owner_id   = betteruptime_monitor.simple.id
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
    email = "petr@betterstack.com" # Replace with your team member's e-mail
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

# Metadata with Team and Schedule references, attached to an incoming webhook
resource "betteruptime_metadata" "escalation_targets" {
  owner_type = "IncomingWebhook"
  owner_id   = betteruptime_incoming_webhook.this.id
  key        = "Escalation targets"
  metadata_value {
    type = "Team"
    name = "Terraform E2E Tests" # Name of one of your teams
  }
  metadata_value {
    type    = "Schedule"
    item_id = betteruptime_on_call_calendar.this.id # Reference an on-call calendar by ID
  }
}
