# All IPs Better Stack checks from

data "betteruptime_ip_list" "this" {}

# IPs to allowlist on your infrastructure

output "betteruptime_ips" {
  value = data.betteruptime_ip_list.this.ips
}

# Available monitoring clusters

output "betteruptime_clusters" {
  value = data.betteruptime_ip_list.this.all_clusters
}

# Only the IPs of selected clusters

data "betteruptime_ip_list" "us" {
  filter_clusters = ["us"]
}

output "betteruptime_us_ips" {
  value = data.betteruptime_ip_list.us.ips
}
