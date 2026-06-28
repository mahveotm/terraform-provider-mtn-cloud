resource "mtncloud_cypher_secret" "db_password" {
  key   = "myapp/db-password"
  value = var.db_password
  ttl   = 0 # 0 = no expiry
}
