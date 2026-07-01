# Creates Jira issues from incidents. Requires a Jira account already connected
# in Better Stack, so this example is documented but excluded from the combined
# E2E run.
resource "betteruptime_jira_integration" "this" {
  name                     = "Terraform Jira Integration"
  automatic_issue_creation = true
  jira_project_key         = "OPS"
  jira_issue_type_id       = "10001"
}
