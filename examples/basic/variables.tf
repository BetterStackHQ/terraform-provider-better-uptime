variable "betteruptime_api_token" {
  type        = string
  description = <<EOF
Better Uptime API Token
(https://docs.betteruptime.com/api/getting-started#obtaining-an-api-token)
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
}
