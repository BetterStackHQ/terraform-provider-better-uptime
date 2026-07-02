resource "betteruptime_heartbeat" "this" {
  name               = "example.com heartbeat"
  period             = 3600
  grace              = 0
  heartbeat_group_id = betteruptime_heartbeat_group.this.id
  # Keep cron-style checks aligned across DST; only applies to periods of 1 hour or longer
  server_timezone      = "Europe/Berlin"
  policy_id            = betteruptime_policy.this.id # Escalate a missed heartbeat through this policy
  team_wait            = 180                         # Wait 3 minutes before escalating to the whole team
  maintenance_from     = "01:00:00"                  # Suppress incidents during a nightly window
  maintenance_to       = "03:00:00"
  maintenance_days     = ["sat", "sun"]
  maintenance_timezone = "Berlin" # Rails timezone name, as the API stores it
}

# Have your job send a request here on each successful run
output "heartbeat_url" {
  value = betteruptime_heartbeat.this.url
}
