# Look up an existing escalation policy by name
data "betteruptime_policy" "existing" {
  name = "My Existing Policy"
}

# The looked-up policy ID, e.g. to attach a monitor to it
output "policy_id" {
  value = data.betteruptime_policy.existing.id
}
