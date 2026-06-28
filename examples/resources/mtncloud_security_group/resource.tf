resource "mtncloud_security_group" "web" {
  name        = "web-servers"
  description = "Allow SSH and HTTPS"
}
