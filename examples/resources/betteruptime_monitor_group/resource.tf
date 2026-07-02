resource "betteruptime_monitor_group" "this" {
  name = "Example monitors"
}

resource "betteruptime_monitor_group" "secondary" {
  name = "Secondary monitors"
  # sort_index orders sibling groups in the dashboard
  sort_index = 2
}
