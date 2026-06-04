#!/usr/bin/env bash
set -euo pipefail

manifest_path="${RELEASE_MANIFEST:-release/manifest/${VERSION:-v0.1.0}.json}"
latest_path="${RELEASE_LATEST_MANIFEST:-release/manifest/latest.json}"
GOWORK=off go run ./internal/tools/releasemanifest --out "$manifest_path"
if [[ "$manifest_path" != "$latest_path" ]]; then
  mkdir -p "$(dirname "$latest_path")"
  cp "$manifest_path" "$latest_path"
  sha256sum "$latest_path" > "${latest_path}.sha256"
fi
