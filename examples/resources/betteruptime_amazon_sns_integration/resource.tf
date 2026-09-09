# An Amazon SNS integration opens an incident for every notification the topic publishes.
#
# Subscribe the topic to the integration's url on the AWS side, with protocol HTTPS and raw
# message delivery disabled. Better Stack answers the SubscriptionConfirmation handshake, then
# fills in topic_arn and subscription_state, which is why both are read-only here.
resource "betteruptime_amazon_sns_integration" "backend_alerts" {
  name = "Terraform Amazon SNS"

  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
  team_wait      = 300

  # "unused" matches every message; "any" and "all" would each need a rules block.
  started_rule_type      = "unused"
  acknowledged_rule_type = "unused"
  resolved_rule_type     = "unused"

  # SNS wraps the payload in an envelope; the inner Message is what carries the alert.
  cause_field {
    name         = "Cause"
    special_type = "cause"
    field_target = "body"
    match_type   = "match_everything"
  }

  # Subject lives on the envelope rather than in the message, so it needs the sns_envelope
  # target. It is optional in the SNS spec: when the publisher omits it the incident keeps the
  # cause as its name rather than an empty title.
  title_field {
    name         = "Title"
    special_type = "title"
    field_target = "sns_envelope"
    target_field = "Subject"
    match_type   = "match_everything"
  }

  started_alert_id_field {
    name         = "Alert ID"
    special_type = "alert_id"
    field_target = "body"
    match_type   = "match_everything"
  }
}

# Subscribe the SNS topic to this URL (HTTPS protocol, raw message delivery disabled)
output "amazon_sns_integration_url" {
  value = betteruptime_amazon_sns_integration.backend_alerts.url
}
