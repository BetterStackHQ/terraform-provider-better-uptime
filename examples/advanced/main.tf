provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

resource "betteruptime_status_page" "this" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  contact_url  = "mailto:support@example.com"
  timezone     = "UTC"
  subdomain    = var.betteruptime_status_page_subdomain
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

resource "betteruptime_monitor" "this" {
  url              = "https://example.com"
  monitor_type     = "status"
  monitor_group_id = betteruptime_monitor_group.this.id
}

resource "betteruptime_status_page_resource" "monitor" {
  status_page_id         = betteruptime_status_page.this.id
  status_page_section_id = betteruptime_status_page_section.monitors.id
  resource_id            = betteruptime_monitor.this.id
  resource_type          = "Monitor"
  public_name            = "example.com site"
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
  name = "Terraform Email integration"
  call = false
  sms = false
  email = true
  push = true
  team_wait = 180
  recovery_period = 0
  paused = false
  started_rule_type = "any"
  acknowledged_rule_type = "unused"
  resolved_rule_type = "all"

  started_rules {
    rule_target = "subject"
    match_type = "contains"
    content = "[Alert]"
  }
  started_rules {
    rule_target = "subject"
    match_type = "contains"
    content = "[Alert Reminder]"
  }
  resolved_rules {
    rule_target = "subject"
    match_type = "contains"
    content = "[Resolved Alert]"
  }

  cause_field {
    name = "Cause"
    special_type = "cause"
    field_target = "subject"
    match_type = "match_everything"
  }
  started_alert_id_field {
    name = "Alert ID"
    special_type = "alert_id"
    field_target = "subject"
    match_type = "match_between"
    content_before = "]"
    content_after = ")"
  }
  resolved_alert_id_field {
    name = "Alert ID"
    special_type = "alert_id"
    field_target = "subject"
    match_type = "match_between"
    content_before = "]"
    content_after = ")"
  }

  other_started_fields {
    name = "Caused by"
    field_target = "body"
    match_type = "match_between"
    content_before = "by:"
    content_after = "\n"
  }
  other_started_fields {
    name = "Description"
    field_target = "body"
    match_type = "match_between"
    content_before = "Description:"
    content_after = "\n"
  }
}

