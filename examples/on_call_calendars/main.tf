provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

data "betteruptime_on_call_calendar" "primary" {}

data "betteruptime_on_call_calendar" "secondary" {
  name = var.betteruptime_secondary_calendar_name
}

resource "betteruptime_on_call_calendar" "new" {
  name = "My Terraform calendar"
  on_call_rotation {
    users              = var.betteruptime_rotation_users
    rotation_length    = 1
    rotation_interval  = "day"
    start_rotations_at = "2025-01-01T00:00:00Z"
    end_rotations_at   = "2026-01-01T00:00:00Z"
  }
}
