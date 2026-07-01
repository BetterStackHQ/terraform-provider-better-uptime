resource "betteruptime_status_page_section" "monitors" {
  status_page_id = betteruptime_status_page.this.id
  name           = "Our monitors"
  position       = 0
}

resource "betteruptime_status_page_section" "heartbeats" {
  status_page_id = betteruptime_status_page.this.id
  name           = "Our heartbeats"
  position       = 1
}

resource "betteruptime_status_page_section" "manually_tracked_items" {
  status_page_id = betteruptime_status_page.this.id
  name           = "Manually tracked items"
  position       = 2
}
