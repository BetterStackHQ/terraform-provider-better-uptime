# Look up an existing Amazon SNS integration by name
data "betteruptime_amazon_sns_integration" "existing" {
  name = "My Existing Amazon SNS Integration"
}

# subscription_state is "active" once AWS has confirmed the subscription, "awaiting" until then
output "existing_amazon_sns_subscription_state" {
  value = data.betteruptime_amazon_sns_integration.existing.subscription_state
}
