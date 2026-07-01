# Invite a team member. A random suffix keeps the invitation e-mail unique per run.
resource "betteruptime_team_member" "new" {
  email = "invitation+${random_pet.unique.id}@example.com"
  role  = "member"
}

# Invite a team member with the built-in team_lead role.
resource "betteruptime_team_member" "lead" {
  email = "tf-lead-${random_pet.unique.id}@example.com" # Unique per run to avoid collisions
  role  = "team_lead"
}
