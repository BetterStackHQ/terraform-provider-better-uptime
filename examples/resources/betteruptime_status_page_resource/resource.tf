# A monitor on the status page
resource "betteruptime_status_page_resource" "monitor_status" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.monitors.id
  resource_id            = betteruptime_monitor.status.id
  resource_type          = "Monitor"
  public_name            = "example.com site"
  widget_type            = "response_times" # Show a response-times chart (also: plain, history, intraday_history)
}

# A heartbeat on the status page
resource "betteruptime_status_page_resource" "heartbeat" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.heartbeats.id
  resource_id            = betteruptime_heartbeat.this.id
  resource_type          = "Heartbeat"
  public_name            = "example.com site (heartbeat)"
  explanation            = "Public API used by customer integrations" # Help text shown next to the resource
  position               = 2                                          # Explicit position within the section, indexed from zero
}

# Publish a whole monitor group - HeartbeatGroup works the same way
resource "betteruptime_status_page_resource" "monitor_group" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.monitors.id
  resource_id            = betteruptime_monitor_group.this.id
  resource_type          = "MonitorGroup"
  public_name            = "All monitors"
}

# A manually tracked item (no backing resource - toggled by hand or via API)
resource "betteruptime_status_page_resource" "manually_tracked_item" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.manually_tracked_items.id
  resource_type          = "ManuallyTrackedItem"
  public_name            = "example.com manual item"
}

# An integration whose up/down state is driven by incident metadata rules
resource "betteruptime_status_page_resource" "email" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.monitors.id
  resource_id            = betteruptime_email_integration.this.id
  resource_type          = "EmailIntegration"
  public_name            = "General status"

  # Mark as down only for incidents matching specific metadata values
  mark_as_down_for = "incident_matching_metadata"
  mark_as_down_metadata_rule {
    key = "Severity"
    metadata_value {
      value = "high"
    }
    metadata_value {
      value = "urgent"
    }
  }

  # Mark as degraded based on User-type metadata
  mark_as_degraded_for = "incident_matching_metadata"
  mark_as_degraded_metadata_rule {
    key = "Assigned User"
    metadata_value {
      type  = "User"
      email = "petr@betterstack.com"
    }
  }
}

# Down whenever there is any ongoing incident
resource "betteruptime_status_page_resource" "down_on_any_incident" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.manually_tracked_items.id
  resource_type          = "ManuallyTrackedItem"
  public_name            = "Down on any incident"
  mark_as_down_for       = "any_incident"
}
