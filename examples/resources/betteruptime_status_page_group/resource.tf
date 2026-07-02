resource "betteruptime_status_page_group" "this" {
  name = "Status pages from Terraform"
}

resource "betteruptime_status_page_group" "secondary" {
  name = "Secondary status pages from Terraform"
  # sort_index orders sibling groups in the dashboard
  sort_index = 2
}
