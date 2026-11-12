resource "betteruptime_on_call_calendar" "this" {
  name = "Terraform on-call calendar"

  # Tags whoever is on call as @terraform-example-on-call in each connected Slack workspace; "" stops tagging
  slack_group_handle = "terraform-example-on-call"
  # true would use a user group that already exists in Slack under that handle instead of creating one
  use_existing_slack_group = false

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

# How each connected Slack workspace took the handle, and who its group can't tag yet
output "betteruptime_on_call_calendar_slack_groups" {
  value = betteruptime_on_call_calendar.this.slack_groups
}
