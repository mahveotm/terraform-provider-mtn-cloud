resource "mtncloud_storage_bucket" "archive" {
  name        = "my-s3-store"
  bucket_name = "my-archive-bucket"
  access_key  = var.s3_access_key
  secret_key  = var.s3_secret_key
  endpoint    = "https://ps1csp-s3.ict.mtn.com.ng:9021"

  create_bucket = true
}

variable "s3_access_key" {
  type      = string
  sensitive = true
}

variable "s3_secret_key" {
  type      = string
  sensitive = true
}
