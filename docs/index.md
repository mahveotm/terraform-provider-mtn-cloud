# MTN Cloud Provider

The MTN Cloud provider manages core MTN Cloud infrastructure through Terraform.

## Provider

```hcl
provider "mtncloud" {
  url      = "https://console.cloud.mtn.ng"
  token    = var.mtncloud_token
  timeout  = 30
  insecure = false

  # Sensible defaults inherited by resources (resource values override these).
  group         = "MTNNG_CLOUD_AZ_1" # default group/site (MTN_CLOUD_GROUP)
  resource_pool = "my-project"       # default pool (MTN_CLOUD_RESOURCE_POOL)

  default_labels = ["terraform"]
  default_tags   = { managed_by = "terraform" }
}
```

### Provider-level defaults

- `group`, `resource_pool`, `availability_zone` — inherited by `mtncloud_instance`
  (and `group`/`resource_pool` by `mtncloud_network`) when the resource omits them.
  Env: `MTN_CLOUD_GROUP`, `MTN_CLOUD_RESOURCE_POOL`, `MTN_CLOUD_AVAILABILITY_ZONE`.
- `default_labels` — union-merged into each resource's `labels` (see `labels_all`).
- `default_tags` — merged into each resource's `tags`, with resource values
  winning per key (see `tags_all`).

Transient API failures (HTTP 429/5xx, network errors) are retried automatically
with exponential backoff; writes (POST/PUT/DELETE) are retried only on 429.

## Data Sources

- `mtncloud_group`
- `mtncloud_instance_type`
- `mtncloud_network`
- `mtncloud_resource_pool`
- `mtncloud_security_group`
- `mtncloud_service_plan`
- `mtncloud_virtual_image`

## Resources

- `mtncloud_archive_bucket`
- `mtncloud_instance`
- `mtncloud_network`
- `mtncloud_security_group`
- `mtncloud_security_group_rule`
- `mtncloud_storage_bucket`
