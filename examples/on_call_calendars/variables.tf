variable "betteruptime_api_token" {
  type        = string
  description = <<EOF
Better Uptime API Token
(https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/#obtaining-an-uptime-api-token)
EOF
  # The value can be omitted if BETTERUPTIME_API_TOKEN env var is set.
  default = null
}

variable "betteruptime_secondary_calendar_name" {
  type        = string
  description = <<EOF
The name of your secondary on-call calendar.
EOF
  default     = "Secondary calendar"
}
