# Minimal heartbeat - a job pings this URL every hour; alerts when the pings stop

resource "betteruptime_heartbeat" "simple" {
  name   = "Simple heartbeat"
  period = 3600
  grace  = 300
}

# Have your job send a request here on each successful run

output "heartbeat_url" {
  value = betteruptime_heartbeat.simple.url
}

resource "betteruptime_heartbeat" "this" {
  name               = "example.com heartbeat"
  period             = 3600
  grace              = 300
  heartbeat_group_id = betteruptime_heartbeat_group.this.id

  # Keep cron-style checks aligned across DST; only applies to periods of 1 hour or longer
  server_timezone = "Europe/Berlin"

  # Escalate a missed heartbeat through this policy
  policy_id = betteruptime_policy.this.id

  # Wait 3 minutes before escalating to the whole team
  team_wait = 180

  # Suppress incidents during a nightly window
  maintenance_from = "01:00:00"
  maintenance_to   = "03:00:00"
  maintenance_days = ["sat", "sun"]
  # Rails timezone name, as the API stores it
  maintenance_timezone = "Berlin"
}
