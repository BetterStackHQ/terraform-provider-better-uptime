# Look up a custom role by name - system role names like member or team_lead work too
data "betteruptime_role" "existing" {
  name = "My custom role"
}

# Assign the role to team members via role_id
output "role_id" {
  value = data.betteruptime_role.existing.id
}
