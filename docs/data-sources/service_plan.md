# mtncloud_service_plan

Looks up a service plan by name/code for a group and instance type.

```hcl
data "mtncloud_service_plan" "small" {
  name  = "G2S4"
  group = "MTNNG_CLOUD_AZ_1"
  type  = "MTN-CS10"
}
```

Computed attributes: `id`, `code`, `max_cpu`, `max_memory`, `max_storage`.
