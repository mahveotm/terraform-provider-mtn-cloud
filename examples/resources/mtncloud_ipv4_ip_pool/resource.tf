resource "mtncloud_ipv4_ip_pool" "app" {
  name    = "app-pool"
  gateway = "10.20.0.1"
  netmask = "255.255.255.0"

  ip_range = [
    {
      starting_address = "10.20.0.10"
      ending_address   = "10.20.0.200"
    }
  ]
}
