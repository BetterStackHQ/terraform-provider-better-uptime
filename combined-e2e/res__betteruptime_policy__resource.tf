# Escalation policy with every step type: escalation (to Slack/webhooks/on-call),
# instructions, time branching, metadata branching, and a wider escalation.
resource "betteruptime_policy" "this" {
  name            = "Terraform Escalation Policy ${random_pet.unique.id}"
  repeat_count    = 3
  repeat_delay    = 60
  policy_group_id = betteruptime_policy_group.this.id

  # After the 3 repeats pass unacknowledged, escalation continues with the fallback
  # policy below instead of stopping - the supported way to keep re-notifying once a
  # policy is exhausted, without escalation loops.
  fallback_policy_id = betteruptime_policy.fallback.id

  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = betteruptime_severity.this.id
    step_members { type = "all_slack_integrations" }
    step_members { type = "all_webhook_integrations" }
    step_members { type = "current_on_call" }
  }
  steps {
    type             = "instructions"
    wait_before      = 0
    reminder_enabled = false
    comment          = <<EOT
# Incident handling instructions

- [ ] Acknowledge the alert
- [ ] Review the incident details and assess impact
- [ ] Document any findings in the incident notes
EOT
  }
  steps {
    type        = "time_branching"
    wait_before = 0
    time_from   = "00:00"
    time_to     = "00:00"
    days        = ["sat", "sun"]
    timezone    = "Eastern Time (US & Canada)"
    policy_id   = null
  }
  steps {
    type         = "metadata_branching"
    wait_before  = 0
    metadata_key = "Description"
    metadata_value {
      value = "Low priority issue"
    }
    metadata_value {
      value = "FYI"
    }
    policy_id = null
  }
  steps {
    type         = "metadata_branching"
    wait_before  = 0
    metadata_key = "Assigned User"
    metadata_value {
      type  = "User"
      email = "petr@betterstack.com"
    }
    policy_metadata_key = "Assigned Policy"
  }
  steps {
    type        = "escalation"
    wait_before = 180
    urgency_id  = betteruptime_severity.this.id
    step_members { type = "entire_team" }
  }
}

# Fallback policy: invoked only after "this" exhausts its repeats unacknowledged,
# widening the blast radius to the whole team.
resource "betteruptime_policy" "fallback" {
  name            = "Terraform Fallback Policy ${random_pet.unique.id}"
  repeat_count    = 5
  repeat_delay    = 120
  policy_group_id = betteruptime_policy_group.this.id

  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = betteruptime_severity.this.id
    step_members { type = "entire_team" }
  }
}

# Silent policy: no steps, so it never alerts anyone - incidents are only collected.
resource "betteruptime_policy" "silent" {
  name            = "Terraform Silent Policy ${random_pet.unique.id}"
  repeat_count    = 0
  repeat_delay    = 0
  policy_group_id = betteruptime_policy_group.this.id
}
