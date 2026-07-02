# A severity (urgency) controls how team members are notified. This one notifies via e-mail and push.
resource "betteruptime_severity" "this" {
  name           = "Terraform Severity ${random_pet.unique.id}" # random_pet keeps names unique when re-running the examples - use a plain name
  email          = true
  push           = true
  call           = false
  sms            = false
  critical_alert = false

  severity_group_id = betteruptime_severity_group.this.id
}
