# Look up an existing monitor by its URL (the only lookup key the schema exposes)
data "betteruptime_monitor" "existing" {
  url = "https://betterstack.com"
}

# Expose the monitor type so callers can confirm what was matched
output "monitor_type" {
  value = data.betteruptime_monitor.existing.monitor_type
}
