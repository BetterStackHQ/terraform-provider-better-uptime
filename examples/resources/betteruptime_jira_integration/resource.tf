# Creates Jira issues from incidents. Requires a Jira account already connected
# in Better Stack, so this example is documented but excluded from the combined
# E2E run
resource "betteruptime_jira_integration" "this" {
  name                     = "Terraform Jira Integration"
  automatic_issue_creation = true
  jira_project_key         = "OPS"
  jira_issue_type_id       = "10001"
  better_stack_id          = "12345"                                                            # ID of the Jira integration to manage - create it in the Better Stack web UI first, then import
  jira_fields_json         = jsonencode({ priority = { id = "3" }, labels = ["better-stack"] }) # Extra Jira fields to set on created issues
}
