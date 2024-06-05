locals {
  primary_on_calls   = data.betteruptime_on_call_calendar.primary.on_call_users
  secondary_on_calls = data.betteruptime_on_call_calendar.secondary.on_call_users
}

output "on_call_primary" {
  value = length(local.primary_on_calls) > 0 ? local.primary_on_calls[0].email : "Nobody on call!"
}

output "on_call_secondary" {
  value = local.secondary_on_calls != null ? (length(local.secondary_on_calls) > 0 ? local.secondary_on_calls[0].email : "Nobody on call!") : "Secondary calendar not found!"
}
