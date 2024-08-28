output "monitoring_ips" {
  value = data.betteruptime_ip_list.this.ips
}

output "monitoring_clusters" {
  value = data.betteruptime_ip_list.this.all_clusters
}
