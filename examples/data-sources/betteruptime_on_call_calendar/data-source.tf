# The team's default on-call calendar (looked up with no arguments)
data "betteruptime_on_call_calendar" "default" {}

# Who is on call right now
output "on_call_user_emails" {
  value = data.betteruptime_on_call_calendar.default.on_call_users[*].email
}

# Look up a specific calendar by name
data "betteruptime_on_call_calendar" "existing" {
  name = "My Existing On-call Calendar"
}

output "existing_on_call_calendar_id" {
  value = data.betteruptime_on_call_calendar.existing.id
}
