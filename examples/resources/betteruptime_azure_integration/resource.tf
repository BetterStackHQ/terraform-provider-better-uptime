# Better Stack receives alerts from Azure through a generated webhook URL.
resource "betteruptime_azure_integration" "this" {
  name           = "Terraform Azure Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}
