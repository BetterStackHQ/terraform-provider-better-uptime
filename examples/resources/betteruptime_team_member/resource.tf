# Invite a team member with the default responder role
resource "betteruptime_team_member" "alice" {
  # Replace with your colleague's e-mail
  email = "alice@betterstack.com"
}

# Invite a team member with the built-in member role
resource "betteruptime_team_member" "bob" {
  email = "bob@betterstack.com"
  role  = "member"
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
