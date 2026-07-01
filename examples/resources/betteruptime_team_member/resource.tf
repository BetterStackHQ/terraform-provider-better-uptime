# Invite a team member. A random suffix keeps the invitation e-mail unique per run.
resource "betteruptime_team_member" "new" {
  email = "invitation+${random_pet.unique.id}@example.com"
  role  = "member"
}
