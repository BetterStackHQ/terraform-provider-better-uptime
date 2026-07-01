resource "betteruptime_heartbeat" "this" {
  name               = "example.com heartbeat"
  period             = 3600
  grace              = 0
  heartbeat_group_id = betteruptime_heartbeat_group.this.id
  # Keep cron-style checks aligned across DST; only applies to periods of 1 hour or longer
  server_timezone = "Europe/Berlin"
}
