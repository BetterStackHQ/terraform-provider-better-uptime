resource "betteruptime_datadog_integration" "this" {
  name           = "Terraform Datadog Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}
