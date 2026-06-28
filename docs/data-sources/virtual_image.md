# mtncloud_virtual_image

Looks up an MTN Cloud virtual image by name.

```hcl
data "mtncloud_virtual_image" "ubuntu" {
  name = "Ubuntu 24.04 LTS"
}
```

## Arguments

- `name` (Required) — Virtual image name.

## Attributes

- `id`, `code`, `image_type`, `os`, `is_public`
