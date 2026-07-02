# Create Jira issues from incidents on demand. Connect the Jira account in the
# Better Stack web UI first, then import the integration here by its better_stack_id

resource "betteruptime_jira_integration" "this" {
  # ID of the Jira integration to manage - create it in the web UI first, the ID is in the integration's URL
  better_stack_id          = "35600"
  name                     = "Terraform Jira Integration"
  automatic_issue_creation = false
  jira_project_key         = "TST"
  jira_issue_type_id       = "10168"

  # Extra Jira fields to set on created issues
  jira_fields_json = jsonencode({
    priority = "3",
    labels = [{
      label = "test"
      value = "test"
    }],
    reporter = {
      label = "Petr Heinz"
      value = "63d92b7086a66a7cc7a54fa8"
    }
  })
}
