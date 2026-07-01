resource "betteruptime_elastic_integration" "this" {
  name           = "Terraform Elastic Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = true
}
