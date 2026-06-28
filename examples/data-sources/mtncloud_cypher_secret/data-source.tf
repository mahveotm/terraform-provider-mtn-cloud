data "mtncloud_cypher_secret" "db_password" {
  key = "myapp/db-password"
}

# Reference the decrypted value via data.mtncloud_cypher_secret.db_password.value
