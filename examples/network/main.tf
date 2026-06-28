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

resource "mtncloud_network" "app" {
  name          = "app-net"
  group         = "MTNNG_CLOUD_AZ_1"
  type          = "openstackPrivate"
  resource_pool = "my-project"

  cidr        = "10.42.10.0/24"
  gateway     = "10.42.10.1"
  dns_primary = "8.8.8.8"
  description = "Application tier network"
  labels      = ["terraform", "app"]
}

# Reference an existing network by name.
data "mtncloud_network" "existing" {
  name = "shared-services"
}
