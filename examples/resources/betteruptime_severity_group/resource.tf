resource "betteruptime_severity_group" "this" {
  name = "Severities from Terraform"
}

resource "betteruptime_severity_group" "secondary" {
  name = "Secondary severities from Terraform"

  # sort_index orders sibling groups in the dashboard
  sort_index = 2
}
