#!/usr/bin/env bash
set -euo pipefail

evidence="${DOWNSTREAM_EVIDENCE:-release/downstream/adoption.json}"

if [[ ! -f "$evidence" ]]; then
  echo "ERROR: missing downstream evidence: $evidence" >&2
  exit 1
fi

required=(
  '"fixtures"'
  '"commands"'
  '"blockers"'
  '"configx"'
  '"corekit"'
  '"external_real_downstream"'
)

for fragment in "${required[@]}"; do
  if ! grep -Fq "$fragment" "$evidence"; then
    echo "ERROR: downstream evidence missing $fragment" >&2
    exit 1
  fi
done

echo "downstream evidence check passed"
