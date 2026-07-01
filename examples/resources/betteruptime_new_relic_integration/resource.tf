# Better Stack receives alerts from New Relic through a generated webhook URL.
resource "betteruptime_new_relic_integration" "this" {
  name           = "Terraform New Relic Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}
