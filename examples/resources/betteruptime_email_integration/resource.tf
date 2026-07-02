# Minimal e-mail integration - open an incident for every inbound e-mail
# (a catch-all started rule matches every subject)
resource "betteruptime_email_integration" "simple" {
  name                   = "Terraform Minimal Email Integration"
  started_rule_type      = "any"
  acknowledged_rule_type = "unused"
  resolved_rule_type     = "unused"

  # Match every e-mail - .* matches any subject, empty included
  started_rules {
    rule_target = "subject"
    match_type  = "matches_regex"
    content     = ".*"
  }

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
}

# Auto-generated inbox to forward alert e-mails to
output "email_integration_address" {
  value = betteruptime_email_integration.simple.email_address
}

# An e-mail integration that parses inbound alert e-mails into incidents, showing
# the full range of field extractors (cause, title, alert id, custom fields)
resource "betteruptime_email_integration" "this" {
  name            = "Terraform Email Integration"
  call            = false
  sms             = false
  email           = true
  push            = true
  critical_alert  = false
  team_wait       = 180
  recovery_period = 0
  paused          = false

  # Route parsed e-mail incidents through this policy
  policy_id = betteruptime_policy.this.id

  started_rule_type = "any"

  # Acknowledge on any matching rule
  acknowledged_rule_type = "any"
  resolved_rule_type     = "all"

  started_rules {
    rule_target = "subject"
    match_type  = "contains"
    content     = "[Alert]"
  }
  started_rules {
    # Only e-mails from your alerting sender open incidents
    rule_target = "from_email"
    match_type  = "equals"
    content     = "alerts@example.com"
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
    # Subject text after the "[Alert]" tag becomes the incident title
    name         = "Title"
    special_type = "title"
    field_target = "subject"
    match_type   = "match_after"

    # match_after and match_before take the marker in content
    content = "]"
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
