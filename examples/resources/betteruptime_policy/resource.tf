# The default High/Low severities exist in every account, so these work out of the box

data "betteruptime_severity" "high" {
  name = "High Severity"
}
data "betteruptime_severity" "low" {
  name = "Low Severity"
}

# Minimal escalation policy - one step that pages whoever is on call

resource "betteruptime_policy" "simple" {
  # random_pet keeps names unique when re-running the examples - use a plain name
  name = "Terraform Simple Policy ${random_pet.unique.id}"

  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = data.betteruptime_severity.high.id
    step_members { type = "current_on_call" }
  }
}

# Escalation policy with repeats and fallback chaining - pages integrations and
# on-call first, then the entire team at 09:00

resource "betteruptime_policy" "this" {
  name            = "Terraform Escalation Policy ${random_pet.unique.id}"
  repeat_count    = 3
  repeat_delay    = 60
  policy_group_id = betteruptime_policy_group.this.id

  # After the 3 repeats pass unacknowledged, escalation continues with the fallback
  # policy below instead of stopping - the supported way to keep re-notifying once a
  # policy is exhausted, without escalation loops
  fallback_policy_id = betteruptime_policy.fallback.id

  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = data.betteruptime_severity.high.id
    step_members { type = "all_slack_integrations" }
    step_members { type = "all_webhook_integrations" }
    step_members { type = "current_on_call" }

    # Chain to the fallback policy
    step_members {
      type = "policy"
      id   = betteruptime_policy.fallback.id
    }
  }
  steps {
    type = "escalation"

    # Run this step at 09:00 local time instead of after a fixed delay
    # (wait_before and wait_until_time are mutually exclusive)
    wait_until_time     = "09:00"
    wait_until_timezone = "Eastern Time (US & Canada)"
    urgency_id          = data.betteruptime_severity.low.id
    step_members { type = "entire_team" }
  }
}

# Post handling instructions, re-remind until they are done, then page whoever is on call

resource "betteruptime_policy" "instructions" {
  name            = "Terraform Instructions Policy ${random_pet.unique.id}"
  policy_group_id = betteruptime_policy_group.this.id

  steps {
    type             = "instructions"
    wait_before      = 0
    reminder_enabled = true

    # Re-remind every 6 hours until the checklist is done
    reminder_interval_hours = 6

    comment = <<EOT
# Incident handling instructions

- [ ] Acknowledge the alert
- [ ] Review the incident details and assess impact
- [ ] Document any findings in the incident notes
EOT
  }
  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = data.betteruptime_severity.high.id
    step_members { type = "current_on_call" }
  }
}

# Route weekend incidents to the silent policy, page whoever is on call otherwise

resource "betteruptime_policy" "business_hours" {
  name            = "Terraform Business Hours Policy ${random_pet.unique.id}"
  policy_group_id = betteruptime_policy_group.this.id

  steps {
    type        = "time_branching"
    wait_before = 0
    days        = ["sat", "sun"]
    time_from   = "00:00"

    # 00:00-00:00 covers the whole day
    time_to  = "00:00"
    timezone = "Eastern Time (US & Canada)"

    # Matching incidents continue with the silent policy below
    policy_id = betteruptime_policy.silent.id
  }
  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = data.betteruptime_severity.high.id
    step_members { type = "current_on_call" }
  }
}

# Route incidents to different policies based on their metadata

resource "betteruptime_policy" "metadata_routing" {
  name            = "Terraform Metadata Routing Policy ${random_pet.unique.id}"
  policy_group_id = betteruptime_policy_group.this.id

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
    metadata_value {
      # Also match incidents whose Description metadata is absent or blank
      type = "no_value"
    }
    policy_id = betteruptime_policy.silent.id
  }
  steps {
    type         = "metadata_branching"
    wait_before  = 0
    metadata_key = "Assigned User"
    metadata_value {
      type = "User"

      # Replace with your team member's e-mail
      email = "petr@betterstack.com"
    }

    # Continue with the policy named by the incident's "Assigned Policy" metadata
    policy_metadata_key = "Assigned Policy"
  }
  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = data.betteruptime_severity.high.id

    # Escalate to whoever the incident metadata names
    step_members {
      type         = "incident_metadata"
      metadata_key = "Assigned Policy"
    }
    step_members { type = "current_on_call" }
  }
}

# Fallback policy: invoked only after "this" exhausts its repeats unacknowledged,
# widening the blast radius to the whole team

resource "betteruptime_policy" "fallback" {
  name            = "Terraform Fallback Policy ${random_pet.unique.id}"
  repeat_count    = 5
  repeat_delay    = 120
  policy_group_id = betteruptime_policy_group.this.id

  steps {
    type        = "escalation"
    wait_before = 0
    urgency_id  = data.betteruptime_severity.high.id
    step_members { type = "entire_team" }
  }
}

# Silent policy: no steps, so it never alerts anyone - incidents are only collected

resource "betteruptime_policy" "silent" {
  name            = "Terraform Silent Policy ${random_pet.unique.id}"
  repeat_count    = 0
  repeat_delay    = 0
  policy_group_id = betteruptime_policy_group.this.id
}
