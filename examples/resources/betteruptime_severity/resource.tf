# A severity (urgency) controls how team members are notified. This one is silent.
resource "betteruptime_severity" "this" {
  name           = "Terraform Severity"
  call           = false
  sms            = false
  email          = false
  push           = false
  critical_alert = false

  severity_group_id = betteruptime_severity_group.this.id
}
