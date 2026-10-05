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
    # Non-default HTTP method
    http_method = "put"

    headers_template {
      name  = "Content-Type"
      value = "application/json"
    }
    # Multiple template headers
    headers_template {
      name = "Authorization"

      # Replace with your endpoint's token
      value = "Bearer your-api-token"
    }

    body_template = jsonencode({ event = "on_call_changed" })
  }
}

# Outgoing webhook fired when a monitor's state changes
resource "betteruptime_outgoing_webhook" "on_monitor_change" {
  name         = "Terraform Monitor Webhook"
  url          = "https://example.com"
  trigger_type = "monitor_change"

  custom_webhook_template_attributes {
    # $METADATA_ARRAY sends the monitor's metadata typed, for example [{"key": "Owner", "values": [{"type": "User", "item_id": 42, "name": "Jane Doe", "email": "jane@example.com"}]}]
    body_template = "{\"monitor\":{\"id\":\"$MONITOR_ID\"},\"metadata\":$METADATA_ARRAY}"
  }
}

# Monitor webhook kept on the legacy metadata format, for a receiver built against the old payload shape
resource "betteruptime_outgoing_webhook" "on_monitor_change_legacy_metadata" {
  name                 = "Terraform Monitor Webhook (legacy metadata)"
  url                  = "https://example.com"
  trigger_type         = "monitor_change"
  metadata_api_version = "v2"

  custom_webhook_template_attributes {
    # On v2, $METADATA_ARRAY is [{"key": "Owner", "value": "jane@example.com"}]: plain text only, one entry per value,
    # and "value": null for team members, teams and other typed values. v3 (the default for new webhooks) sends every value typed.
    body_template = "{\"monitor\":{\"id\":\"$MONITOR_ID\"},\"metadata\":$METADATA_ARRAY}"
  }
}
