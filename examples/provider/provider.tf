terraform {
  required_providers {
    mtncloud = {
      source  = "mahveotm/mtncloud"
      version = "~> 0.1"
    }
  }
}

# Set group/resource_pool once (like AWS's region), plus org-wide tags/labels.
# Resources inherit these unless they override them.
provider "mtncloud" {
  token         = var.mtncloud_token
  group         = "MTNNG_CLOUD_AZ_1"
  resource_pool = "my-project"

  default_labels = ["terraform"]
  default_tags = {
    managed_by = "terraform"
    team       = "platform"
  }
}

variable "mtncloud_token" {
  type      = string
  sensitive = true
}
