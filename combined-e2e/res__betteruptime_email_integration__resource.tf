# An e-mail integration that parses inbound alert e-mails into incidents, showing
# the full range of field extractors (cause, title, alert id, custom fields).
resource "betteruptime_email_integration" "this" {
  name                   = "Terraform Email integration"
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
    rule_target = "subject"
    match_type  = "contains"
    content     = "[Alert]"
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
  title_field {
    name           = "Title"
    special_type   = "title"
    field_target   = "subject"
    match_type     = "match_between"
    content_before = "["
    content_after  = "]"
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
    name           = "Description"
    field_target   = "body"
    match_type     = "match_between"
    content_before = "Description:"
    content_after  = "\n"
  }
  other_started_fields {
    name           = "Severity"
    field_target   = "body"
    match_type     = "match_between"
    content_before = "Severity:"
    content_after  = "\n"
  }
}
