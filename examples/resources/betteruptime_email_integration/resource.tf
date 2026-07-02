# Minimal e-mail integration: open an incident for every inbound e-mail.
# started_rule_type = "all" with no started_rules matches every e-mail
resource "betteruptime_email_integration" "simple" {
  started_rule_type      = "all"
  acknowledged_rule_type = "unused"
  resolved_rule_type     = "unused"

  # Subject becomes the incident title
  title_field {
    name         = "Title"
    special_type = "title"
    field_target = "subject"
    match_type   = "match_everything"
  }
  # E-mail body becomes the incident cause
  cause_field {
    name         = "Cause"
    special_type = "cause"
    field_target = "body"
    match_type   = "match_everything"
  }
  # Alert id is required to dedupe/resolve incidents; take it from the body
  started_alert_id_field {
    name         = "Alert ID"
    special_type = "alert_id"
    field_target = "body"
    match_type   = "match_everything"
  }
}

# Auto-generated inbox to forward alert e-mails to
output "email_integration_address" {
  value = betteruptime_email_integration.simple.email_address
}

# An e-mail integration that parses inbound alert e-mails into incidents, showing
# the full range of field extractors (cause, title, alert id, custom fields)
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
  policy_id              = betteruptime_policy.this.id # Route parsed e-mail incidents through this policy
  started_rule_type      = "any"
  acknowledged_rule_type = "any" # Acknowledge on any matching rule
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
  acknowledged_rules {
    # Match acknowledgement e-mails by subject
    rule_target = "subject"
    match_type  = "contains"
    content     = "[Acknowledged]"
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
  acknowledged_alert_id_field {
    # Extract the alert id when acknowledging
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
  other_started_fields {
    # Extract the error message from the subject with a regex
    name         = "Error message"
    field_target = "subject"
    match_type   = "match_regex"
    content      = "Error: (.*)"
  }

  other_acknowledged_fields {
    name         = "Note"
    field_target = "body"
    match_type   = "match_everything"
  }
  other_resolved_fields {
    name         = "Resolution"
    field_target = "body"
    match_type   = "match_everything"
  }
}
