# Look up an existing severity (urgency) by name
data "betteruptime_severity" "existing" {
  name = "My Existing Severity"
}

# The looked-up severity ID, e.g. to reference from an escalation policy step
output "severity_id" {
  value = data.betteruptime_severity.existing.id
}
