# mtncloud_security_group

Looks up an MTN Cloud security group by name.

```hcl
data "mtncloud_security_group" "default" {
  name = "default"
}
```

## Arguments

- `name` (Required) — Security group name.

## Attributes

- `id`, `description`, `active`
