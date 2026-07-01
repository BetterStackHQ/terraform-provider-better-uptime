resource "betteruptime_on_call_calendar" "this" {
  name = "Terraform on-call calendar"

  on_call_rotation {
    users              = ["petr@betterstack.com"]
    rotation_length    = 1
    rotation_interval  = "day"
    start_rotations_at = "2025-01-01T00:00:00Z"
    end_rotations_at   = "2026-01-01T00:00:00Z"
  }
}
