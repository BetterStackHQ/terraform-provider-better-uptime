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
    type = "User"

    # Replace with your team member's e-mail
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

    # Reference a policy by name instead of by id
    name = "My Existing Escalation Policy"
  }
}

# A Team-typed metadata value, attached to an incoming webhook
# (all values of one metadata key must share the same type)

resource "betteruptime_metadata" "owning_team" {
  owner_type = "IncomingWebhook"
  owner_id   = betteruptime_incoming_webhook.this.id
  key        = "Owning team"
  metadata_value {
    type = "Team"

    # Name of one of your teams
    name = "Terraform E2E Tests"
  }
}

# A Schedule-typed metadata value referencing an on-call calendar

resource "betteruptime_metadata" "escalation_calendar" {
  owner_type = "IncomingWebhook"
  owner_id   = betteruptime_incoming_webhook.this.id
  key        = "Escalation calendar"
  metadata_value {
    type    = "Schedule"
    item_id = betteruptime_on_call_calendar.this.id
  }
}
