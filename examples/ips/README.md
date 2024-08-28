This directory contains a sample Terraform configuration for getting all IPs used by Better Stack for Uptime monitoring.
These IPs can be used to configure resources in other providers accordingly.

## Usage

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime/examples/ips

echo '# See variables.tf for more.
betteruptime_api_token             = "XXXXXXXXXXXXXXXXXXXXXXXX"
betteruptime_clusters              = ["us"]
' > terraform.tfvars

terraform init
terraform apply
```
