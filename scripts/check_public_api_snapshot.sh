#!/usr/bin/env bash
set -euo pipefail

snapshot="contracts/public_api.snapshot"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

GOWORK=off go run ./internal/tools/apisnapshot ./pkg/observex > "$tmp"

if ! diff -u "$snapshot" "$tmp"; then
  echo "ERROR: public API snapshot drift; update contracts/public_api.snapshot intentionally" >&2
  exit 1
fi

echo "public API snapshot check passed"
