# Better Stack posts incident webhooks to a Splunk On-Call (VictorOps) URL

resource "betteruptime_splunk_oncall_integration" "this" {
  name = "Terraform Splunk On-Call Integration"

  # From your Splunk On-Call generic webhook settings
  url = "https://alert.victorops.com/integrations/generic/20131114/alert/your-api-key/your-routing-key"

  # Don't also notify Splunk On-Call when no escalation policy is set (default is true)
  notify_alongside_primary_responder = false
}
