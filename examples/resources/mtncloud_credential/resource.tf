resource "mtncloud_credential" "registry" {
  type     = "username-password"
  name     = "docker-registry"
  username = "ci-bot"
  password = var.registry_password
}

resource "mtncloud_credential" "object_store" {
  type       = "access-key-secret"
  name       = "s3-backups"
  access_key = var.s3_access_key
  secret_key = var.s3_secret_key
}
