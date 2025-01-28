This directory contains a sample Terraform configuration for an example catalog.

## Usage

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime/examples/catalog

echo '# See variables.tf for more.
betteruptime_api_token             = "XXXXXXXXXXXXXXXXXXXXXXXX"
' > terraform.tfvars

terraform init
terraform apply
```
