# Changelog

All notable changes to this provider are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-28

Initial release of the MTN Cloud Terraform provider.

### Added

- **Provider configuration** with OAuth (`username`/`password`) or `token`
  authentication, a configurable API `url`, request `timeout`, and `max_retries`.
- Provider-level defaults `group`, `resource_pool`, and `availability_zone` that
  resources inherit unless overridden (resource value wins, AWS-style).
- `default_labels` and `default_tags` merged into every resource via computed
  `labels_all` / `tags_all` so shared metadata applies without per-resource repetition.
- **Resources**
  - `mtncloud_instance` — provisions instances from human-friendly names (group,
    resource pool, instance type, service plan, image) resolved to IDs internally.
  - `mtncloud_network` — manages networks; group/zone/type/resource-pool given by name.
  - `mtncloud_security_group` and `mtncloud_security_group_rule`.
  - `mtncloud_storage_bucket` — S3-compatible bucket; `secret_key` is write-only
    (the API never returns it).
  - `mtncloud_archive_bucket` — archive bucket backed by a storage provider.
- **Data sources** — `mtncloud_group`, `mtncloud_resource_pool`,
  `mtncloud_service_plan`, `mtncloud_instance_type`, `mtncloud_virtual_image`,
  `mtncloud_network`, `mtncloud_security_group`.
- **Plan-time validation** — CIDR blocks, port ranges, protocol/direction/policy
  enums, visibility and retention-policy values, VLAN range, and positive
  day/timeout values.
- Automatic retry with exponential backoff and jitter (429 on any method; 5xx and
  network errors on GETs only) honoring `Retry-After`.
- Import support for all resources via `terraform import`.

[Unreleased]: https://github.com/mahveotm/terraform-provider-mtncloud/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mahveotm/terraform-provider-mtncloud/releases/tag/v0.1.0
