#!/usr/bin/env bash
# check-redaction-leak.sh — CI script to detect potential secret leaks in logs.
#
# Scans non-test .go files for patterns that may leak sensitive data:
#   1. Missing use of SecretString/Sanitize/RedactField for sensitive values
#   2. Direct formatting of secret-bearing variables without redaction
#
# Exit code 0: no violations found.
# Exit code 1: violations found.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VIOLATIONS=0

while IFS= read -r -d '' file; do
  # Skip test files, vendor, testdata, and worktrees.
  case "$file" in
    */vendor/*|*/testdata/*|*/.worktree/*|*_test.go) continue ;;
  esac

  # Find lines that log sensitive keywords without redaction helpers.
  while IFS= read -r match; do
    lineno="${match%%:*}"
    content="${match#*:}"

    # Skip if the line uses any redaction mechanism.
    if echo "$content" | grep -qE '(SecretString|RedactField|RedactFields|RedactedValue|IsSecretKey|\.Redact|Sanitize|Secret\()' 2>/dev/null; then
      continue
    fi

    # Skip comments, variable/constant declarations, and function definitions.
    if echo "$content" | grep -qE '(var |const |//|/\*|func |DefaultDenied)' 2>/dev/null; then
      continue
    fi

    echo "VIOLATION: $file:$lineno: sensitive value may be logged without redaction"
    echo "  $content"
    echo "  Use SecretString, Secret(), or RedactField() before logging."
    VIOLATIONS=$((VIOLATIONS + 1))
  done < <(grep -nE '(fmt\.(Print|Sprint|Fprint|Sprintf|Printf)|\.Info|\.Error|\.Warn|\.Debug)\b.*\b(password|secret|token|api_key|access_key|secret_key)\b' "$file" 2>/dev/null \
    | grep -v 'SecretString' \
    | grep -v 'RedactField' \
    | grep -v 'RedactedValue' \
    | grep -v 'Sanitize' \
    | grep -v 'IsSecretKey' \
    | grep -v 'Secret(' \
    | grep -v '^\s*//' \
    | grep -v 'func ' \
    | grep -v 'var ' \
    | grep -v 'DefaultDenied' \
    || true)
done < <(find "$REPO_ROOT" -name '*.go' -not -path '*/vendor/*' -not -path '*/.worktree/*' -print0)

if [[ $VIOLATIONS -gt 0 ]]; then
  echo ""
  echo "Found $VIOLATIONS potential redaction leak(s)."
  echo "All secret/sensitive values must be wrapped with SecretString, Secret(),"
  echo "or passed through RedactField/RedactFields before reaching log output."
  exit 1
fi

echo "check-redaction-leak: no potential secret leaks detected."
exit 0
