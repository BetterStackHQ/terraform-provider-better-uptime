# Better Stack receives alerts from CloudWatch through a generated webhook URL
resource "betteruptime_aws_cloudwatch_integration" "this" {
  name           = "Terraform CloudWatch Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}

resource "betteruptime_aws_cloudwatch_integration" "with_policy" {
  name      = "Terraform CloudWatch Integration with custom policy"
  policy_id = betteruptime_policy.this.id
}

# Point CloudWatch at this URL to deliver alerts to Better Stack
output "aws_cloudwatch_integration_webhook_url" {
  value = betteruptime_aws_cloudwatch_integration.this.webhook_url
}
