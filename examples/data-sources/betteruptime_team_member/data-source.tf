# Look up a team member by e-mail
data "betteruptime_team_member" "existing" {
  # Replace with your team member's e-mail
  email = "petr@betterstack.com"
}

output "existing_team_member_id" {
  value = data.betteruptime_team_member.existing.id
}
