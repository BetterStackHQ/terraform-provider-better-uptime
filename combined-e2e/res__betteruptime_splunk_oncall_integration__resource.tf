# Better Stack posts incident webhooks to a Splunk On-Call (VictorOps) URL
resource "betteruptime_splunk_oncall_integration" "this" {
  name = "Terraform Splunk On-Call Integration"
  url  = "https://alert.victorops.com/integrations/generic/00000000/alert/example"

  notify_alongside_primary_responder = false # Don't also notify Splunk On-Call when no escalation policy is set (default is true)
}
