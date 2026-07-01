# Look up a monitor you manage by its URL
data "betteruptime_monitor" "existing" {
  url        = betteruptime_monitor.status.url
  depends_on = [betteruptime_monitor.status]
}
