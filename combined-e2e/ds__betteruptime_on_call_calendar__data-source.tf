# The team's default on-call calendar (looked up with no arguments).
data "betteruptime_on_call_calendar" "default" {}

output "on_call_users" {
  value = data.betteruptime_on_call_calendar.default.on_call_users
}
