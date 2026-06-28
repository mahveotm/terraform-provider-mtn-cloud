# mtncloud_resource_pool

Looks up a resource pool by name or code within a group.

```hcl
data "mtncloud_resource_pool" "project" {
  name  = "my-project"
  group = "MTNNG_CLOUD_AZ_1"
}
```

Computed attributes: `id`, `code`.
