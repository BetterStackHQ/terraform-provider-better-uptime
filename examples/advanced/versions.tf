terraform {
  required_version = ">= 0.14"
  required_providers {
    betteruptime = {
      source  = "BetterStackHQ/better-uptime"
      version = ">= 0.19.8"
    }
  }
}
