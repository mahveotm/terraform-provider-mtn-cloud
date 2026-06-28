# mtncloud_network

Looks up an MTN Cloud network by name.

```hcl
data "mtncloud_network" "shared" {
  name = "shared-services"
}
```

## Arguments

- `name` (Required) — Network name.
- `cloud_id` (Optional) — Cloud/zone ID to disambiguate networks with the same
  name. Obtain it from `mtncloud_group.cloud_ids`.

## Attributes

- `id`, `code`, `cidr`, `gateway`, `status`, `type_id`, `cloud_id`
