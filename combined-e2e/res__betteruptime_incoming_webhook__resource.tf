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
  started_rule_type      = "any"
  acknowledged_rule_type = "unused"
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
    name         = "Severity"
    field_target = "query_string"
    target_field = "severity"
    match_type   = "match_everything"
  }
}
