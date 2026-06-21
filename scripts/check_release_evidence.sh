#!/usr/bin/env bash
set -euo pipefail

release_version="${VERSION:-}"
if [[ -z "$release_version" ]]; then
  echo "ERROR: VERSION is required; set VERSION=vX.Y.Z" >&2
  exit 1
fi
if [[ ! "$release_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?$ ]]; then
  echo "ERROR: VERSION must look like vX.Y.Z, got: $release_version" >&2
  exit 1
fi
package_version="$(sed -nE 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"([^"]+)".*/\1/p' pkg/observex/version.go | head -n1)"
if [[ -z "$package_version" ]]; then
  echo "ERROR: could not determine package version from pkg/observex/version.go" >&2
  exit 1
fi
if [[ "$package_version" != "$release_version" ]]; then
  echo "ERROR: VERSION $release_version does not match pkg/observex/version.go ($package_version)" >&2
  exit 1
fi
export VERSION="$release_version"

manifest_path="${RELEASE_MANIFEST:-release/manifest/${release_version}.json}"
latest_path="${RELEASE_LATEST_MANIFEST:-release/manifest/latest.json}"
args=(--verify "$manifest_path" --expect-version "$release_version")

if [[ "${RELEASE_EVIDENCE_REQUIRE_PASSED:-0}" == "1" ]]; then
  args+=(--require-passed)
fi

if [[ "${RELEASE_EVIDENCE_REQUIRE_CLEAN:-0}" == "1" ]]; then
  args+=(--require-clean)
fi

./scripts/check_downstream_evidence.sh

GOWORK=off go run ./internal/tools/releasemanifest "${args[@]}"

verify_sidecar() {
  local artifact="$1"
  local sidecar="$2"
  if [[ ! -f "$artifact" ]]; then
    echo "missing release artifact: $artifact" >&2
    exit 1
  fi
  if [[ ! -f "$sidecar" ]]; then
    echo "missing release artifact sha256: $sidecar" >&2
    exit 1
  fi

  local got
  local want
  got="$(awk '{print $1}' "$sidecar")"
  want="$(sha256sum "$artifact" | awk '{print $1}')"
  if [[ "$got" != "$want" ]]; then
    echo "sha256 mismatch for $artifact: got $got want $want" >&2
    exit 1
  fi
}

verify_sidecar "$manifest_path" "${manifest_path}.sha256"
verify_sidecar "$latest_path" "${latest_path}.sha256"
if ! cmp -s "$manifest_path" "$latest_path"; then
  echo "latest manifest drift: $latest_path must match $manifest_path byte-for-byte" >&2
  exit 1
fi

require_doc_marker() {
  local file="$1"
  local marker="$2"
  if [[ ! -f "$file" ]]; then
    echo "missing release evidence document: $file" >&2
    exit 1
  fi
  if ! grep -Fq "$marker" "$file"; then
    echo "release evidence document $file missing marker: $marker" >&2
    exit 1
  fi
}

require_doc_marker "docs/evidence.md" "Public API signature snapshot"
require_doc_marker "docs/evidence.md" "Memory-canonical testkit"
require_doc_marker "docs/downstream-evidence.md" "external_real_downstream"
require_doc_marker "docs/release.md" "release-final-check"
require_doc_marker "docs/structural-analysis-2026-06-04.md" "本地结构分 100/100"

echo "release evidence check passed"
