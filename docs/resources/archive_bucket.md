# mtncloud_archive_bucket

Manages an MTN Cloud archive bucket backed by a storage provider (a
`mtncloud_storage_bucket`, referenced by name).

```hcl
resource "mtncloud_archive_bucket" "vault" {
  name             = "project-vault"
  storage_provider = "my-s3-store"
  description      = "Long-term archive for project artifacts"
  visibility       = "private"
}
```

## Arguments

- `name` (Required) — Globally unique archive bucket name.
- `storage_provider` (Required, ForceNew) — Storage bucket name that backs this archive bucket.
- `description` (Optional)
- `visibility` (Optional) — `private` or `public`.
- `is_public` (Optional, bool)
- `account_id` (Optional, number)

## Attributes

- `id`, `code`, `file_count`, `raw_size`

## Import

```bash
terraform import mtncloud_archive_bucket.vault 12
```
