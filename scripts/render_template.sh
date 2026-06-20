#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/render_template.sh --module-name NAME --module-path PATH --package-name NAME --out DIR

Renders observex into a concrete base library by copying the repository,
moving pkg/observex to pkg/<package>, and replacing template identifiers.
The current source token requires --module-name and --package-name to match.
USAGE
}

module_name=""
module_path=""
package_name=""
out_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module-name)
      module_name="${2:-}"
      shift 2
      ;;
    --module-path)
      module_path="${2:-}"
      shift 2
      ;;
    --package-name)
      package_name="${2:-}"
      shift 2
      ;;
    --out)
      out_dir="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$module_name" || -z "$module_path" || -z "$package_name" || -z "$out_dir" ]]; then
  echo "ERROR: --module-name, --module-path, --package-name and --out are required" >&2
  usage >&2
  exit 2
fi

if [[ "$package_name" =~ [^a-zA-Z0-9_] || "$package_name" =~ ^[0-9] ]]; then
  echo "ERROR: --package-name must be a valid Go package identifier" >&2
  exit 2
fi

if [[ "$module_name" != "$package_name" ]]; then
  echo "ERROR: --module-name and --package-name must match while observex is a shared source token" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$out_dir"
if find "$out_dir" -mindepth 1 -maxdepth 1 | read -r _; then
  echo "ERROR: output directory must be empty: $out_dir" >&2
  exit 2
fi

(
  cd "$repo_root"
  tar \
    --exclude='./.git' \
    --exclude='./.omx' \
    --exclude='./.omc' \
    --exclude='./.worktree' \
    --exclude='./release/manifest/v*.json' \
    -cf - .
) | (
  cd "$out_dir"
  tar -xf -
)

if [[ "$package_name" != "observex" ]]; then
  mkdir -p "$out_dir/pkg"
  mv "$out_dir/pkg/observex" "$out_dir/pkg/$package_name"
fi

replace_in_text_files() {
    local find_text="$1"
    local replace_text="$2"

    while IFS= read -r -d '' file; do
      FIND_TEXT="$find_text" REPLACE_TEXT="$replace_text" perl -0pi -e 's/\Q$ENV{FIND_TEXT}\E/$ENV{REPLACE_TEXT}/g' "$file"
  done < <(
    find "$out_dir" -type f \( \
      -name '*.go' -o \
      -name '*.md' -o \
      -name '*.json' -o \
      -name '*.snapshot' -o \
      -name '*.txt' -o \
      -name '*.sh' -o \
      -name '*.yml' -o \
      -name '*.yaml' -o \
      -name '*.txt' -o \
      -name 'Makefile' -o \
      -name 'go.mod' \
    \) -print0
  )
}

# Replace the most-specific module path before the shared project/package token.
# Keep module_name/package_name equal until explicit placeholders exist.
replace_in_text_files 'github.com/ZoneCNH/observex' "$module_path"
replace_in_text_files 'observex' "$module_name"

(
  cd "$out_dir"
  gofmt -w ./pkg ./internal ./contracts ./examples ./testkit
)

echo "rendered $module_name at $out_dir"
