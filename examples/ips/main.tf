provider "betteruptime" {
  api_token = var.betteruptime_api_token
}

data "betteruptime_ip_list" "this" {
  filter_clusters = var.betteruptime_clusters
}
