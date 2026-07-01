# Outgoing webhook fired on incident changes, with a custom request template
resource "betteruptime_outgoing_webhook" "on_incident" {
  name         = "Terraform Outgoing Webhook"
  url          = "https://example.com"
  trigger_type = "incident_change"

  on_incident_started      = true
  on_incident_acknowledged = false
  on_incident_resolved     = false
  on_incident_reopened     = false
  on_incident_comment      = false

  notify_alongside_primary_responder = false

  custom_webhook_template_attributes {
    http_method = "get"

    auth_username = "user"
    auth_password = "password"

    headers_template {
      name  = "Content-Type"
      value = "application/json"
    }

    body_template = jsonencode({
      incident = {
        id         = "$INCIDENT_ID"
        started_at = "$STARTED_AT"
      }
    })
  }
}

# Outgoing webhook fired when the on-call shift changes
resource "betteruptime_outgoing_webhook" "on_call_change" {
  name         = "Terraform On-call Webhook"
  url          = "https://example.com"
  trigger_type = "on_call_change"

  custom_webhook_template_attributes {
    http_method = "put" # Non-default HTTP method

    headers_template {
      name  = "Content-Type"
      value = "application/json"
    }
    headers_template {
      # Multiple template headers
      name  = "Authorization"
      value = "Bearer $INCIDENT_ID"
    }

    body_template = "{\"incident\":{\"id\":\"$INCIDENT_ID\",\"started_at\":\"$STARTED_AT\"}}"
  }
}

# Outgoing webhook fired when a monitor's state changes
resource "betteruptime_outgoing_webhook" "on_monitor_change" {
  name         = "Terraform Monitor Webhook"
  url          = "https://example.com"
  trigger_type = "monitor_change" # Fire when a monitor's state changes

  custom_webhook_template_attributes {
    body_template = "{\"incident\":{\"id\":\"$INCIDENT_ID\",\"started_at\":\"$STARTED_AT\"}}"
  }
}
