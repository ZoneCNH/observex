#!/usr/bin/env bash
set -euo pipefail

echo "checking secrets..."

PATTERNS=(
  "password="
  "passwd="
  "secret="
  "token="
  "access_key="
  "secret_key="
  "AKIA[0-9A-Z]{16}"
  "BEGIN RSA PRIVATE KEY"
  "BEGIN OPENSSH PRIVATE KEY"
)

for pattern in "${PATTERNS[@]}"; do
  if grep -R -E "$pattern" . \
    --exclude-dir=.git \
    --exclude-dir=.omx \
    --exclude-dir=vendor \
    --exclude="*.sum" \
    --exclude="check_secrets.sh" \
    --exclude="goal.md" \
    --exclude="*_test.go"; then
    # *_test.go 排除：脱敏/边界测试 fixture 合法包含密钥模式串（如
    # {"password=secret123", true} 用于验证 FR-005/BR-007 脱敏逻辑），非真实凭证。
    echo "ERROR: possible secret found: $pattern"
    exit 1
  fi
done

echo "secret check passed"
