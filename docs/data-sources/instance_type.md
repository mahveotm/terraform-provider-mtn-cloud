# mtncloud_instance_type

Looks up an MTN Cloud instance type by code.

```hcl
data "mtncloud_instance_type" "ubuntu" {
  code = "MTN-CS10"
}
```

Computed attributes: `id`, `name`, `default_layout_id`.
