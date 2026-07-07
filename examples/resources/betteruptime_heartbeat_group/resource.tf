resource "betteruptime_heartbeat_group" "this" {
  name = "Example heartbeats"
}

resource "betteruptime_heartbeat_group" "secondary" {
  name = "Secondary heartbeats"

  # sort_index orders sibling groups in the dashboard
  sort_index = 2
}
