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

# An archive bucket is backed by an existing storage bucket (the storage provider).
resource "mtncloud_archive_bucket" "vault" {
  name             = "project-vault"
  storage_provider = "my-s3-store"
  description      = "Long-term archive for project artifacts"
  visibility       = "private"
}
