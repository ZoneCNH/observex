#!/usr/bin/env bash
set -euo pipefail

release_version="${VERSION:-}"
if [[ -z "$release_version" ]]; then
  release_version="$(sed -nE 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"([^"]+)".*/\1/p' pkg/observex/version.go | head -n1)"
fi
if [[ -z "$release_version" ]]; then
  echo "ERROR: could not determine release version; set VERSION=vX.Y.Z" >&2
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
