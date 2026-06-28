# mtncloud_instance

Manages an MTN Cloud instance using human-friendly provisioning inputs.

With `group` and `resource_pool` set on the provider, a minimal instance is just
name, type, and plan:

```hcl
resource "mtncloud_instance" "web" {
  name = "web-01"
  type = "MTN-CS10"
  plan = "G2S4"

  labels = ["web"]
  tags   = { role = "web" }
}
```

Fully specified:

```hcl
resource "mtncloud_instance" "web" {
  name          = "web-01"
  group         = "MTNNG_CLOUD_AZ_1"
  type          = "MTN-CS10"
  plan          = "G2S4"
  resource_pool = "my-project"

  security_group = "web-servers"
  labels         = ["web"]
}
```

## Arguments

- `name` (Required), `type` (Required), `plan` (Required).
- `group` (Optional) — defaults to the provider's `group`.
- `resource_pool` (Optional) — defaults to the provider's `resource_pool`; if
  neither is set and the group has exactly one pool, that pool is used. If the
  group has several pools, the error lists them.
- `availability_zone` (Optional) — defaults to the provider's `availability_zone`.
- `security_group` (Optional) — defaults to `"default"`.
- Also: `description`, `environment`, `labels`, `tags`, `security_groups`,
  `os_external_network_id`, `create_user`, `workflow_id`, `shutdown_days`,
  `expire_days`, `create_backup`, `wait_for_ready`, `timeouts`.

## Attributes

- `labels_all` — effective labels (provider `default_labels` ∪ `labels`).
- `tags_all` — effective tags (provider `default_tags` overlaid by `tags`).
- `status`, `primary_ip`, `external_ip`, `cloud_id`, `group_id`, `layout_id`,
  `plan_id`, `resource_pool_id`.

`group`, `type`, `resource_pool`, and `availability_zone` force replacement when
changed.

## Import

```bash
terraform import mtncloud_instance.web 123
```
