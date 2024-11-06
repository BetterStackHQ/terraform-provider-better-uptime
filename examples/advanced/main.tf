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
  contact_url  = "mailto:support@example.com"
  timezone     = "Eastern Time (US & Canada)"
  subdomain    = coalesce(var.betteruptime_status_page_subdomain, random_id.status_page_subdomain.hex)
  subscribable = true
}

resource "betteruptime_status_page_section" "monitors" {
  status_page_id = betteruptime_status_page.this.id
  name           = "Our monitors"
  position       = 0
}

resource "betteruptime_status_page_section" "heartbeats" {
  status_page_id = betteruptime_status_page.this.id
  name           = "Our heartbeats"
  position       = 1
}

resource "betteruptime_monitor_group" "this" {
  name       = "example"
  sort_index = 0
}

resource "betteruptime_monitor" "status" {
  url              = "https://example.com"
  monitor_type     = "status"
  monitor_group_id = betteruptime_monitor_group.this.id
  request_headers = [
    {
      "name" : "X-For-Status-Page",
      "value" : "https://${betteruptime_status_page.this.subdomain}.betteruptime.com"
    },
  ]
}

resource "betteruptime_monitor" "dns" {
  url              = "1.1.1.1"
  monitor_type     = "dns"
  request_body     = "example.com"
  monitor_group_id = betteruptime_monitor_group.this.id
}

resource "betteruptime_status_page_resource" "monitor_status" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.monitors.id
  resource_id            = betteruptime_monitor.status.id
  resource_type          = "Monitor"
  public_name            = "example.com site"
}

resource "betteruptime_status_page_resource" "monitor_dns" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.monitors.id
  resource_id            = betteruptime_monitor.dns.id
  resource_type          = "Monitor"
  public_name            = "example.com domain"
}

resource "betteruptime_heartbeat_group" "this" {
  name       = "example"
  sort_index = 0
}

resource "betteruptime_heartbeat" "this" {
  name               = "example.com heartbeat"
  period             = 30
  grace              = 0
  heartbeat_group_id = betteruptime_heartbeat_group.this.id
}

resource "betteruptime_status_page_resource" "heartbeat" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.heartbeats.id
  resource_id            = betteruptime_heartbeat.this.id
  resource_type          = "Heartbeat"
  public_name            = "example.com site (heartbeat)"
}

resource "betteruptime_email_integration" "this" {
  name                   = "Terraform Email integration"
  call                   = false
  sms                    = false
  email                  = true
  push                   = true
  team_wait              = 180
  recovery_period        = 0
  paused                 = false
  started_rule_type      = "any"
  acknowledged_rule_type = "unused"
  resolved_rule_type     = "all"

  started_rules {
    rule_target = "subject"
    match_type  = "contains"
    content     = "[Alert]"
  }
  started_rules {
    rule_target = "subject"
    match_type  = "contains"
    content     = "[Alert Reminder]"
  }
  resolved_rules {
    rule_target = "subject"
    match_type  = "contains"
    content     = "[Resolved Alert]"
  }

  cause_field {
    name         = "Cause"
    special_type = "cause"
    field_target = "subject"
    match_type   = "match_everything"
  }
  started_alert_id_field {
    name           = "Alert ID"
    special_type   = "alert_id"
    field_target   = "subject"
    match_type     = "match_between"
    content_before = "]"
    content_after  = ")"
  }
  resolved_alert_id_field {
    name           = "Alert ID"
    special_type   = "alert_id"
    field_target   = "subject"
    match_type     = "match_between"
    content_before = "]"
    content_after  = ")"
  }

  other_started_fields {
    name           = "Caused by"
    field_target   = "body"
    match_type     = "match_between"
    content_before = "by:"
    content_after  = "\n"
  }
  other_started_fields {
    name           = "Description"
    field_target   = "body"
    match_type     = "match_between"
    content_before = "Description:"
    content_after  = "\n"
  }
}

resource "betteruptime_incoming_webhook" "this" {
  name                   = "Terraform Incoming Webhook"
  call                   = false
  sms                    = false
  email                  = true
  push                   = true
  team_wait              = 180
  recovery_period        = 0
  paused                 = false
  started_rule_type      = "any"
  acknowledged_rule_type = "unused"
  resolved_rule_type     = "all"

  started_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "alert"
  }
  started_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "reminder"
  }
  resolved_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "resolved"
  }

  cause_field {
    field_target = "json"
    target_field = "incident.status"
    match_type   = "match_everything"
  }
  title_field {
    field_target = "json"
    target_field = "incident.title"
    match_type   = "match_everything"
  }
  started_alert_id_field {
    name           = "Alert ID"
    special_type   = "alert_id"
    field_target   = "json"
    target_field   = "incident.id"
    match_type     = "match_between"
    content_before = "<"
    content_after  = "-"
  }
  resolved_alert_id_field {
    name           = "Alert ID"
    special_type   = "alert_id"
    field_target   = "json"
    target_field   = "incident.id"
    match_type     = "match_between"
    content_before = "<"
    content_after  = "-"
  }

  other_started_fields {
    name           = "Caused by"
    field_target   = "json"
    target_field   = "incident.description"
    match_type     = "match_between"
    content_before = "by:"
    content_after  = ","
  }
  other_started_fields {
    name           = "Description"
    field_target   = "json"
    target_field   = "incident.description"
    match_type     = "match_between"
    content_before = "description:"
    content_after  = ","
  }
  other_started_fields {
    name         = "Severity"
    field_target = "query_string"
    target_field = "severity"
    match_type   = "match_everything"
  }
}

resource "betteruptime_policy_group" "this" {
  name = "Policies from Terraform"
}

resource "betteruptime_policy" "this" {
  name            = "Standard Escalation Policy"
  repeat_count    = 3
  repeat_delay    = 60
  policy_group_id = betteruptime_policy_group.this.id

  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = betteruptime_severity.this.id
    step_members { type = "all_slack_integrations" }
    step_members { type = "all_webhook_integrations" }
    step_members { type = "current_on_call" }
  }
  steps {
    type        = "time_branching"
    wait_before = 0
    time_from   = "00:00"
    time_to     = "00:00"
    days        = ["sat", "sun"]
    timezone    = "Eastern Time (US & Canada)"
    policy_id   = null
  }
  steps {
    type            = "metadata_branching"
    wait_before     = 0
    metadata_key    = "Description"
    metadata_values = ["Low priority issue", "FYI", "Notice"]
    policy_id       = null
  }
  steps {
    type        = "escalation"
    wait_before = 180
    urgency_id  = betteruptime_severity.this.id
    step_members { type = "entire_team" }
  }
}

resource "betteruptime_severity_group" "this" {
  name = "Severities from Terraform"
}

resource "betteruptime_severity" "this" {
  name  = var.betteruptime_severity_name
  call  = false
  email = false
  push  = false
  sms   = false

  severity_group_id = betteruptime_severity_group.this.id
}

resource "betteruptime_outgoing_webhook" "this" {
  name         = "Terraform Outgoing Webhook"
  url          = "https://example.com"
  trigger_type = "incident_change"

  on_incident_started      = true
  on_incident_acknowledged = false
  on_incident_resolved     = false

  custom_webhook_template_attributes {
    http_method = "get"

    auth_user     = "user"
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
