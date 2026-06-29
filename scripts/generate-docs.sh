#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

provider_name="mtncloud"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

go run ./tools/docschema -out "${tmp_dir}/schema.json"

go tool tfplugindocs generate \
  --provider-name "${provider_name}" \
  --providers-schema "${tmp_dir}/schema.json"
