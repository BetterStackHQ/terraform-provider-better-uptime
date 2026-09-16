resource "betteruptime_on_call_calendar" "this" {
  name = "Terraform on-call calendar"

  on_call_rotation {
    # Replace with your team members' e-mails
    users             = ["petr@betterstack.com"]
    rotation_length   = 1
    rotation_interval = "day"
    # Midnight in the rotation's time zone; the offset fixes the first shift, the zone keeps that wall clock across daylight-saving changes
    start_rotations_at = "2025-01-01T00:00:00+01:00"
    end_rotations_at   = "2030-01-01T00:00:00Z"
    timezone           = "Europe/Prague"

    # One block per day and window
    working_hours {
      day        = "monday"
      start_time = "09:00"
      end_time   = "17:00"
    }

    # Omit the times for a whole day
    working_hours {
      day = "saturday"
    }

    # Several blocks for one day are allowed
    working_hours {
      day        = "friday"
      start_time = "09:00"
      end_time   = "12:00"
    }
    working_hours {
      day        = "friday"
      start_time = "12:30"
      end_time   = "17:00"
    }
  }
}
