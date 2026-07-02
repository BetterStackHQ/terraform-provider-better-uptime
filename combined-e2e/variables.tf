variable "betteruptime_api_token" {
  type        = string
  description = <<EOF
Better Stack Uptime API token
(https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/#get-an-api-token)
EOF
  # The value can be omitted if the BETTERUPTIME_API_TOKEN env var is set
  default = null
}

variable "betteruptime_status_page_subdomain" {
  type        = string
  description = "Subdomain for the status page. A random one is used when omitted."
  default     = null
}
