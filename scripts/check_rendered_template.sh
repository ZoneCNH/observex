#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/check_rendered_template.sh DIR MODULE_NAME MODULE_PATH PACKAGE_NAME

Checks that a rendered template has no stale template identifiers and exposes
the expected Go module and package directory.
USAGE
}

if [[ $# -ne 4 ]]; then
  usage >&2
  exit 2
fi

repo_dir="$1"
module_name="$2"
module_path="$3"
package_name="$4"

if [[ ! -d "$repo_dir" ]]; then
  echo "ERROR: rendered directory does not exist: $repo_dir" >&2
  exit 2
fi

if [[ "$module_name" != "$package_name" ]]; then
  echo "ERROR: MODULE_NAME and PACKAGE_NAME must match while observex is a shared source token" >&2
  exit 2
fi

actual_module="$(cd "$repo_dir" && GOWORK=off go list -m)"
if [[ "$actual_module" != "$module_path" ]]; then
  echo "ERROR: module path mismatch: got $actual_module, want $module_path" >&2
  exit 1
fi

if [[ ! -d "$repo_dir/pkg/$package_name" ]]; then
  echo "ERROR: rendered package directory missing: pkg/$package_name" >&2
  exit 1
fi

if [[ "$package_name" != "observex" && -e "$repo_dir/pkg/observex" ]]; then
  echo "ERROR: stale pkg/observex directory still exists" >&2
  exit 1
fi

scan_regex() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden --glob '!.git/**' --glob '!**/.foundationx-baseline.txt' "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if grep -RInE --exclude-dir=.git --exclude='*.foundationx-baseline.txt' "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_fixed() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden --glob '!.git/**' --glob '!**/.foundationx-baseline.txt' --fixed-strings "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if grep -RInF --exclude-dir=.git --exclude='*.foundationx-baseline.txt' "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_regex '\{\{MODULE_NAME\}\}|\{\{MODULE_PATH\}\}|\{\{PACKAGE_NAME\}\}' "template placeholder"
scan_fixed "github.com/ZoneCNH/observex" "module path"

legacy_template_name="baselib"'-template'
legacy_template_path="github.com/ZoneCNH/${legacy_template_name}"
scan_fixed "$legacy_template_path" "legacy standard source module path"
scan_fixed "$legacy_template_name" "legacy standard source name"

if [[ "$module_name" != "observex" ]]; then
  scan_fixed "observex" "module name"
fi

if [[ "$package_name" != "observex" ]]; then
  scan_regex '\bobservex\b' "package name"
fi

echo "rendered template check passed: $module_name"
