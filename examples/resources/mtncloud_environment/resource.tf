resource "mtncloud_environment" "production" {
  name        = "Production"
  code        = "prod"
  description = "Production deployment environment"
  visibility  = "private"
  active      = true
}
