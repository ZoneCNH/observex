#!/usr/bin/env bash
set -euo pipefail

echo "checking contracts..."

REQUIRED_FILES=(
  "contracts/config.schema.json"
  "contracts/health.schema.json"
  "contracts/error.schema.json"
  "contracts/field.schema.json"
  "contracts/logger.schema.json"
  "contracts/tracer.schema.json"
  "contracts/metrics.schema.json"
  "contracts/metrics.md"
  "contracts/metric_naming.md"
  "contracts/public_api.md"
  "contracts/public_api.snapshot"
)

for file in "${REQUIRED_FILES[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: missing contract file: $file"
    exit 1
  fi
done

./scripts/check_public_api_snapshot.sh
GOWORK=off go test ./contracts

echo "contract check passed"
