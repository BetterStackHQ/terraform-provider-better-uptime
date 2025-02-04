This directory contains a sample Terraform configuration for extracting current on-call
user's email address for your primary on-call calendar, and a secondary one (fetched by name)

## Usage

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime/examples/on_call_calendars

echo '# See variables.tf for more.
betteruptime_api_token               = "XXXXXXXXXXXXXXXXXXXXXXXX"
betteruptime_secondary_calendar_name = "My secondary calendar"
betteruptime_rotation_users = ["albert@acme.com", "betty@acme.com", "charlie@acme.com"]
' > terraform.tfvars

terraform init
terraform apply
```
