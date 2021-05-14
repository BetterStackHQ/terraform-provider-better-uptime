terraform {
  required_version = ">= 0.13"
  required_providers {
    betteruptime = {
      source = "altinity/betteruptime"
      # https://github.com/altinity/terraform-provider-betteruptime/blob/master/CHANGELOG.md
      version = ">= 0.1.0"
    }
  }
}
