# Minimal incoming webhook - open an incident for every inbound call
# (a catch-all started rule matches every request body)
resource "betteruptime_incoming_webhook" "simple" {
  name = "Terraform Minimal Incoming Webhook"

  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  team_wait      = 300

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
  recovery_period = 0
  paused          = false

  # Route webhook incidents through this policy
  policy_id = betteruptime_policy.this.id

  started_rule_type = "any"

  # Acknowledge on any matching rule
  acknowledged_rule_type = "any"
  resolved_rule_type     = "all"

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

  started_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "alert"
  }

  # Rules can also match request headers
  started_rules {
    rule_target  = "header"
    target_field = "X-Alert"
    match_type   = "matches_regex"
    content      = "firing"
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
  other_started_fields {
    name         = "Severity"
    field_target = "query_string"
    target_field = "severity"
    match_type   = "match_everything"
  }

  # Fields can be extracted from request headers too
  other_started_fields {
    name         = "Alert header"
    field_target = "header"
    target_field = "X-Alert"
    match_type   = "match_everything"
  }

  acknowledged_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "acknowledged"
  }
  acknowledged_alert_id_field {
    name         = "Alert ID"
    special_type = "alert_id"
    field_target = "json"
    target_field = "incident.id"
    match_type   = "match_everything"
  }
  other_acknowledged_fields {
    name         = "Ack note"
    field_target = "json"
    target_field = "incident.note"
    match_type   = "match_everything"
  }

  resolved_rules {
    rule_target  = "json"
    target_field = "incident.status"
    match_type   = "contains"
    content      = "resolved"
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
  other_resolved_fields {
    name         = "Resolver"
    field_target = "json"
    target_field = "incident.resolver"
    match_type   = "match_everything"
  }
}
