terraform {
  required_version = ">= 0.13"
  required_providers {
    betteruptime = {
      source = "BetterStackHQ/betteruptime"
      # https://github.com/BetterStackHQ/terraform-provider-betteruptime/blob/master/CHANGELOG.md
      version = ">= 0.2.0"
    }
  }
}
