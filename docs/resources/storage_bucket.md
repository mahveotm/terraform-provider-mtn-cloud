# mtncloud_storage_bucket

Manages an MTN Cloud S3-compatible storage bucket.

```hcl
resource "mtncloud_storage_bucket" "archive" {
  name        = "my-s3-store"
  bucket_name = "my-archive-bucket"
  access_key  = var.s3_access_key
  secret_key  = var.s3_secret_key
  endpoint    = "https://ps1csp-s3.ict.mtn.com.ng:9021"

  create_bucket = true
}
```

## Arguments

- `name` (Required) — Unique storage bucket name in MTN Cloud.
- `bucket_name` (Required) — Backing S3 bucket name.
- `access_key` (Required, Sensitive)
- `secret_key` (Required, Sensitive) — The API never returns it, so changes are
  applied but drift cannot be detected. Rotate by changing the value here.
- `endpoint` (Required) — S3-compatible endpoint URL.
- `create_bucket` (Optional, default `true`) — Create the backing bucket if missing.
- `storage_server` (Optional, number)
- `default_backup_target`, `copy_to_store`, `default_deployment_target`, `default_virtual_image_target` (Optional, bool)
- `retention_policy_type` (Optional) — `none`, `backup`, or `delete`.
- `retention_policy_days` (Optional, number)
- `retention_provider` (Optional)

## Attributes

- `id`

## Import

```bash
terraform import mtncloud_storage_bucket.archive 7
```

Note: `secret_key` is not returned by the API and will be empty after import;
set it in configuration to manage the bucket going forward.
