# Monitors are looked up by their URL

data "betteruptime_monitor" "existing" {
  url = "https://betterstack.com"
}

# Expose the monitor type so callers can confirm what was matched

output "monitor_type" {
  value = data.betteruptime_monitor.existing.monitor_type
}
