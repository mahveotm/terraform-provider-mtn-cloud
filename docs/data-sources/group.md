# mtncloud_group

Looks up an MTN Cloud group/site by name.

```hcl
data "mtncloud_group" "az1" {
  name = "MTNNG_CLOUD_AZ_1"
}
```

Computed attributes: `id`, `cloud_ids`, `location`, `active`.
