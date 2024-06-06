terraform {
  required_version = ">= 0.13"
  required_providers {
    betteruptime = {
      source = "BetterStackHQ/better-uptime"
      # https://github.com/BetterStackHQ/terraform-provider-better-uptime/blob/master/CHANGELOG.md
      version = ">= 0.8.0"
    }
  }
}
