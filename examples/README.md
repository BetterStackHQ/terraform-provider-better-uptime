# Example usage
[![build](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/build.yml)
[![tests](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/test.yml)
[![documentation](https://img.shields.io/badge/-documentation-blue)](https://registry.terraform.io/providers/BetterStackHQ/better-uptime/latest/docs)

These examples demonstrate how to provision and manage Better Stack Uptime resources such as
monitors, heartbeats, status pages, on-call calendars and integrations with Terraform.

The files in this directory are a ready-to-run **basic** example - a status page with a monitor.

Detailed, per-resource examples live in [`resources/`](./resources) and per-data-source examples in
[`data-sources/`](./data-sources) - these are embedded in the registry docs and exercised
end-to-end in CI.

## Usage

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime/examples

echo '# See variables.tf for more.
betteruptime_api_token = "XXXXXXXXXXXXXXXXXXXXXXXX"
' > terraform.tfvars

terraform init
terraform apply

# open the status page created above
open $(terraform output -raw betteruptime_status_page_url)
```

## Documentation

See [Better Stack Uptime API docs](https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/) to obtain an API token and the complete list of parameter options.
Or explore the [Terraform Registry provider documentation](https://registry.terraform.io/providers/BetterStackHQ/better-uptime/latest/docs).
