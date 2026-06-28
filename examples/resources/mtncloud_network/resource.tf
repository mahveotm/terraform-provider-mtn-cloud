resource "mtncloud_network" "app" {
  name          = "app-net"
  group         = "MTNNG_CLOUD_AZ_1"
  type          = "openstackPrivate"
  resource_pool = "my-project"

  cidr        = "10.42.10.0/24"
  gateway     = "10.42.10.1"
  dns_primary = "8.8.8.8"
  description = "Application tier network"
  labels      = ["terraform", "app"]
}
