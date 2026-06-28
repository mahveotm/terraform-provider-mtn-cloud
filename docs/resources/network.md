# mtncloud_network

Manages an MTN Cloud network. References to the group, zone, network type, and
resource pool are given as human-friendly names/codes and resolved to IDs
automatically.

```hcl
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
```

## Arguments

- `name` (Required) — Network name.
- `group` (Optional, ForceNew) — Group/site name. Defaults to the provider's
  `group`. The group's first cloud is used as the network's zone (see `cloud_id`).
- `type` (Optional, ForceNew) — Network type name or code (e.g. an OpenStack network type).
- `resource_pool` (Optional, ForceNew) — Resource pool name or code. Defaults to
  the provider's `resource_pool`. Required for OpenStack networks.
- `cidr`, `gateway`, `dns_primary`, `dns_secondary`, `description`, `visibility` (Optional)
- `vlan_id` (Optional, number)
- `dhcp_server`, `assign_public_ip`, `allow_static_override`, `active` (Optional, bool)
- `labels` (Optional, list of string)

## Attributes

- `labels_all` — effective labels (provider `default_labels` ∪ `labels`).

- `id`, `code`, `status`
- `cloud_id`, `group_id`, `type_id`, `resource_pool_id`

## Import

```bash
terraform import mtncloud_network.app 99
```
