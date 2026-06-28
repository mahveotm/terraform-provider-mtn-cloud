terraform {
  required_providers {
    mtncloud = {
      source  = "mahveotm/mtn-cloud"
      version = "0.1.0"
    }
  }
}

provider "mtncloud" {
  token = var.mtncloud_token
}

variable "mtncloud_token" {
  type      = string
  sensitive = true
}
