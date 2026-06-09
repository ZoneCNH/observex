#!/usr/bin/env bash
# check-label-policy.sh — CI script to enforce metric label naming policy.
#
# Scans non-test .go files for metric label patterns and checks:
#   1. Label keys in Labels{...} map literals match lower snake_case
#   2. No denied label names are used as metric label keys (error, err, msg, level)
#
# Exit code 0: no violations found.
# Exit code 1: violations found.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DENIED_LABELS=("error" "err" "msg" "level")
VIOLATIONS=0

while IFS= read -r -d '' file; do
  # Skip test files, vendor, testdata, and worktrees.
  case "$file" in
    */vendor/*|*/testdata/*|*/.worktree/*|*_test.go) continue ;;
  esac

  # Look for Labels{...} map literals with string keys.
  # Pattern: "key": value  or  "key": value,
  while IFS= read -r match; do
    lineno="${match%%:*}"
    content="${match#*:}"

    # Extract string keys from map literals (Labels{"key": ...}).
    while [[ "$content" =~ \"([a-zA-Z_][a-zA-Z0-9_]*)\":  ]]; do
      label="${BASH_REMATCH[1]}"
      content="${content#*\"${label}\"}"

      # Skip long strings and URLs.
      if [[ ${#label} -gt 40 ]] || [[ "$label" == */* ]] || [[ "$label" == http* ]]; then
        continue
      fi

      # Check denied labels.
      for denied in "${DENIED_LABELS[@]}"; do
        if [[ "$label" == "$denied" ]]; then
          echo "VIOLATION: $file:$lineno: denied metric label \"$label\" — use error_kind or status instead"
          VIOLATIONS=$((VIOLATIONS + 1))
        fi
      done

      # Check snake_case for lowercase-starting labels that look like metric keys.
      if [[ "$label" =~ ^[a-z] ]] && ! [[ "$label" =~ ^[a-z][a-z0-9_]*$ ]]; then
        echo "VIOLATION: $file:$lineno: metric label \"$label\" is not lower snake_case"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
    done
  done < <(grep -n '"[a-zA-Z_][a-zA-Z0-9_]*":\s' "$file" 2>/dev/null || true)
done < <(find "$REPO_ROOT" -name '*.go' -not -path '*/vendor/*' -not -path '*/.worktree/*' -print0)

if [[ $VIOLATIONS -gt 0 ]]; then
  echo ""
  echo "Found $VIOLATIONS label policy violation(s)."
  echo "Denied labels (${DENIED_LABELS[*]}) must not be used as metric label keys."
  echo "Use error_kind instead of error/err for error classification."
  exit 1
fi

echo "check-label-policy: all metric labels pass policy."
exit 0
