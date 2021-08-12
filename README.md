# terraform-provider-better-uptime [![build](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/BetterStackHQ/terraform-provider-better-uptime/actions/workflows/build.yml) [![documentation](https://img.shields.io/badge/-documentation-blue)](https://registry.terraform.io/providers/BetterStackHQ/better-uptime/latest/docs)

Terraform (0.13+) provider for [Better Uptime](https://betteruptime.com/).

## Installation

```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    betteruptime = {
      source = "BetterStackHQ/better-uptime"
      # https://github.com/BetterStackHQ/terraform-provider-better-uptime/blob/master/CHANGELOG.md
      version = ">= 0.2.0"
    }
  }
}
```

## Example Usage

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

> See [examples/](examples/) for more.

## Documentation

See Terraform Registry [docs](https://registry.terraform.io/providers/BetterStackHQ/better-uptime/latest/docs).

## Development

> PREREQUISITE: [go1.16+](https://golang.org/dl/).

```shell script
git clone https://github.com/BetterStackHQ/terraform-provider-better-uptime && \
  cd terraform-provider-better-uptime

make help
```
