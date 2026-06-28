resource "mtncloud_network_domain" "corp" {
  name        = "corp.example.ng"
  fqdn        = "corp.example.ng"
  description = "Corporate DNS domain"
  visibility  = "private"
  active      = true
}
