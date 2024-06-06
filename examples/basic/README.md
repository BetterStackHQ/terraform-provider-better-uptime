This directory contains a sample Terraform configuration for a status page with a single monitor.

## Usage

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime/examples/basic

echo '# See variables.tf for more.
betteruptime_api_token             = "XXXXXXXXXXXXXXXXXXXXXXXX"
betteruptime_status_page_subdomain = "example"
' > terraform.tfvars

terraform init
terraform apply

# open https://example.betteruptime.com (or different, based on variable above)
open $(terraform output -raw betteruptime_status_page_url)
```
