#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

provider_name="mtncloud"
provider_source="registry.terraform.io/mahveotm/mtncloud"
provider_version="0.0.1"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

goos="$(go env GOOS)"
goarch="$(go env GOARCH)"
plugin_dir="${tmp_dir}/plugins/${provider_source}/${provider_version}/${goos}_${goarch}"
mkdir -p "${plugin_dir}" "${tmp_dir}/work"

binary="${plugin_dir}/terraform-provider-${provider_name}_v${provider_version}"
if [[ "${goos}" == "windows" ]]; then
  binary="${binary}.exe"
fi

go build -o "${binary}" .

cat >"${tmp_dir}/terraformrc" <<EOF
provider_installation {
  dev_overrides {
    "${provider_source}" = "${plugin_dir}"
  }
  direct {}
}
EOF

cat >"${tmp_dir}/work/main.tf" <<EOF
terraform {
  required_providers {
    ${provider_name} = {
      source  = "mahveotm/${provider_name}"
      version = "${provider_version}"
    }
  }
}

provider "${provider_name}" {}
EOF

TF_CLI_CONFIG_FILE="${tmp_dir}/terraformrc" \
  terraform -chdir="${tmp_dir}/work" providers schema -json >"${tmp_dir}/schema.json"

go run ./tools/docschemaalias \
  -in "${tmp_dir}/schema.json" \
  -out "${tmp_dir}/schema-short.json" \
  -source "${provider_source}" \
  -alias "${provider_name}"

go tool tfplugindocs generate \
  --provider-name "${provider_name}" \
  --providers-schema "${tmp_dir}/schema-short.json"
