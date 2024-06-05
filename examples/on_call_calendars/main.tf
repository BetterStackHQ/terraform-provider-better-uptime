provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

data "betteruptime_on_call_calendar" "primary" {}

data "betteruptime_on_call_calendar" "secondary" {
  name = var.betteruptime_secondary_calendar_name
}
