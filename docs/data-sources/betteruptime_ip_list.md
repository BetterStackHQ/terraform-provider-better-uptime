---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "betteruptime_ip_list Data Source - terraform-provider-better-uptime"
subcategory: ""
description: |-
  Monitoring IPs lookup.
---

# betteruptime_ip_list (Data Source)

Monitoring IPs lookup.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter_clusters` (List of String) The optional list of clusters used to filter the IPs.

### Read-Only

- `all_clusters` (List of String) The list of all clusters.
- `id` (String) The internal ID of the resource, can be ignored.
- `ips` (List of String) The list of IPs used for monitoring.


