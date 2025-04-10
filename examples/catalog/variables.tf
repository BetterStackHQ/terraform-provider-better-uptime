variable "betteruptime_api_token" {
  type        = string
  description = <<EOF
Better Stack Uptime API token
(https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/#obtaining-an-uptime-api-token)
EOF
  # The value can be omitted if BETTERUPTIME_API_TOKEN env var is set.
  default = null
}
