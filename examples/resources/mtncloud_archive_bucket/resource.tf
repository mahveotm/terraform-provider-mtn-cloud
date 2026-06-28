# An archive bucket is backed by an existing storage bucket (the storage provider).
resource "mtncloud_archive_bucket" "vault" {
  name             = "project-vault"
  storage_provider = "my-s3-store"
  description      = "Long-term archive for project artifacts"
  visibility       = "private"
}
