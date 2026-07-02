# Invite a team member with the built-in member role
resource "betteruptime_team_member" "alice" {
  email = "alice@betterstack.com" # Replace with your colleague's e-mail
  role  = "member"
}

# Invite a team member with the built-in team_lead role
resource "betteruptime_team_member" "bob" {
  email = "bob@betterstack.com"
  role  = "team_lead"
}

# Look up a custom role by name to assign it by id
data "betteruptime_role" "custom" {
  name = "My custom role"
}

# Invite a team member with a custom role via role_id (set only one of role or role_id)
resource "betteruptime_team_member" "dylan" {
  email   = "dylan@betterstack.com"
  role_id = data.betteruptime_role.custom.id
}
