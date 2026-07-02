# All Better Stack monitoring IP ranges, e.g. to allowlist them on your firewall.
data "betteruptime_ip_list" "this" {}

output "betteruptime_ips" {
  value = data.betteruptime_ip_list.this.all_clusters
}
