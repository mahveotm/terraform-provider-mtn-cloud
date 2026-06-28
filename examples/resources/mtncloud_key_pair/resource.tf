resource "mtncloud_key_pair" "deploy" {
  name       = "deploy"
  public_key = file("~/.ssh/id_rsa.pub")
}
