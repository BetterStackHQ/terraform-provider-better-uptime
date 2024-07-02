variable "betteruptime_api_token" {
  type        = string
  description = <<EOF
Better Stack Uptime API Token
(https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/#obtaining-an-uptime-api-token)
EOF
  # The value can be omitted if BETTERUPTIME_API_TOKEN env var is set.
  default = null
}

variable "betteruptime_status_page_subdomain" {
  type        = string
  description = <<EOF
betteruptime.com status page subdomain
(e.g. if you set value to "my-status-page" your status page will be
available at https://my-status-page.betteruptime.com)
EOF
  # The value can be omitted and random domain will be provided.
  default = null
}

variable "betteruptime_severity_name" {
  type        = string
  description = "Name of the severity from Better Uptime you want to use with Escalation policies created using Terraform"
  default     = "Low Severity"
}
