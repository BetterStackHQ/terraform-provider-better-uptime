# Minimal incoming webhook - open an incident for every inbound call
# (a catch-all started rule matches every request body)

resource "betteruptime_incoming_webhook" "simple" {
  name                   = "Terraform Minimal Incoming Webhook"
  started_rule_type      = "any"
  acknowledged_rule_type = "unused"
  resolved_rule_type     = "unused"

  # Match every request - .* matches any body, empty included
  started_rules {
    rule_target = "body"
    match_type  = "matches_regex"
    content     = ".*"
  }

  # Whole request body becomes the incident title
  title_field {
    name         = "Title"
    special_type = "title"
    field_target = "body"
    match_type   = "match_everything"
  }

  # Whole request body becomes the incident cause
  cause_field {
    name         = "Cause"
    special_type = "cause"
    field_target = "body"
    match_type   = "match_everything"
  }
}

# Auto-generated URL to POST alert payloads to

output "incoming_webhook_url" {
  value = betteruptime_incoming_webhook.simple.url
}

# An incoming webhook that parses JSON/query-string payloads into incidents

resource "betteruptime_incoming_webhook" "this" {
  name            = "Terraform Incoming Webhook"
  call            = false
  sms             = false
  email           = true
  push            = true
  critical_alert  = false
  team_wait       = 180
  recovery_period = 0
  paused          = false

  # Route webhook incidents through this policy
  policy_id = betteruptime_policy.this.id

  started_rule_type      = "any"
  acknowledged_rule_type = "any"
  resolved_rule_type     = "all"

  started_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "alert"
  }
  resolved_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "resolved"
  }
  acknowledged_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "acknowledged"
  }
  started_rules {
    # Rules can also match request headers
    rule_target  = "header"
    target_field = "X-Alert"
    match_type   = "matches_regex"
    content      = "firing"
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
  acknowledged_alert_id_field {
    name         = "Alert ID"
    special_type = "alert_id"
    field_target = "json"
    target_field = "incident.id"
    match_type   = "match_everything"
  }

  other_started_fields {
    name         = "Severity"
    field_target = "query_string"
    target_field = "severity"
    match_type   = "match_everything"
  }
  other_started_fields {
    # Fields can be extracted from request headers too
    name         = "Alert header"
    field_target = "header"
    target_field = "X-Alert"
    match_type   = "match_everything"
  }
  other_acknowledged_fields {
    name         = "Ack note"
    field_target = "json"
    target_field = "incident.note"
    match_type   = "match_everything"
  }
  other_resolved_fields {
    name         = "Resolver"
    field_target = "json"
    target_field = "incident.resolver"
    match_type   = "match_everything"
  }
}
