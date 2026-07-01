# Better Stack receives alerts from CloudWatch through a generated webhook URL.
resource "betteruptime_aws_cloudwatch_integration" "this" {
  name           = "Terraform CloudWatch Integration"
  call           = false
  sms            = false
  email          = true
  push           = true
  critical_alert = false
}
