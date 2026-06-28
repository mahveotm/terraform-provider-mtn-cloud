terraform {
  required_providers {
    mtncloud = {
      source  = "mahveotm/mtn-cloud"
      version = "0.1.0"
    }
  }
}

# Set the group/resource_pool once on the provider (like AWS's region), plus
# org-wide tags/labels. Resources inherit these unless they override them.
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

# Minimal instance: group, resource_pool, and security_group ("default") are
# inherited/defaulted; labels/tags merge with the provider defaults.
resource "mtncloud_instance" "web" {
  name = "web-01"
  type = "MTN-CS10"
  plan = "G2S4"

  labels = ["web"]            # effective labels_all = ["terraform", "web"]
  tags   = { role = "web" }   # effective tags_all merges provider default_tags
}
