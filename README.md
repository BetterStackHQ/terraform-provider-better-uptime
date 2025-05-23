# terraform-provider-better-uptime
[![build](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/build.yml)
[![tests](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/test.yml)
[![documentation](https://img.shields.io/badge/-documentation-blue)](https://registry.terraform.io/providers/BetterStackHQ/better-uptime/latest/docs)

Terraform (0.14+) provider for [Better Stack Uptime](https://uptime.betterstack.com/).

## Installation

```terraform
terraform {
  required_version = ">= 0.14"
  required_providers {
    betteruptime = {
      source  = "BetterStackHQ/better-uptime"
      version = ">= 0.9.3"
    }
  }
}
```

## Example usage

See [`/examples` directory](./examples) for multiple ready-to-use examples.
Here's a simple one to get you started:

```terraform
provider "betteruptime" {
  # `api_token` can be omitted if BETTERUPTIME_API_TOKEN env var is set.
  api_token = "XXXXXXXXXXXXXXXXXXXXXXXX"
}

resource "betteruptime_status_page" "this" {
  company_name = "Example, Inc"
  company_url  = "https://example.com"
  timezone     = "UTC"
  subdomain    = "example"
}

resource "betteruptime_monitor" "this" {
  url          = "https://example.com"
  monitor_type = "status"
}

resource "betteruptime_status_page_resource" "monitor" {
  status_page_id = betteruptime_status_page.this.id
  resource_id    = betteruptime_monitor.this.id
  resource_type  = "Monitor"
  public_name    = "example.com site"
}
```

## Documentation

See [Better Stack Uptime API docs](https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/) to obtain API token and get the complete list of parameter options.
Or explore the [Terraform Registry provider documentation](https://registry.terraform.io/providers/BetterStackHQ/better-uptime/latest/docs).

## Development

> PREREQUISITE: [go1.23+](https://golang.org/dl/).

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime

make help
```
