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
GOWORK=off go run ./internal/tools/releasemanifest --out "$manifest_path"
if [[ "$manifest_path" != "$latest_path" ]]; then
  mkdir -p "$(dirname "$latest_path")"
  cp "$manifest_path" "$latest_path"
  sha256sum "$latest_path" > "${latest_path}.sha256"
fi
