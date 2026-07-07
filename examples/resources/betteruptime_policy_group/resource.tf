resource "betteruptime_policy_group" "this" {
  name = "Policies from Terraform"
}

resource "betteruptime_policy_group" "secondary" {
  name = "Secondary policies from Terraform"

  # sort_index orders sibling groups in the dashboard
  sort_index = 2
}
