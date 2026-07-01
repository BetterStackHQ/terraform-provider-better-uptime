# An incoming webhook that parses JSON/query-string payloads into incidents.
resource "betteruptime_incoming_webhook" "this" {
  name                   = "Terraform Incoming Webhook"
  call                   = false
  sms                    = false
  email                  = true
  push                   = true
  critical_alert         = false
  team_wait              = 180
  recovery_period        = 0
  paused                 = false
  policy_id              = betteruptime_policy.this.id # Route webhook incidents through this policy
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
    # Cover header target and regex rule matching
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
    # Cover the header field target
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
