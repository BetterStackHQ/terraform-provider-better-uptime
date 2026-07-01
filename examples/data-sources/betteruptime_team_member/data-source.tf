data "betteruptime_team_member" "existing" {
  email = "petr@betterstack.com"
}

output "existing_team_member_id" {
  value = data.betteruptime_team_member.existing.id
}
