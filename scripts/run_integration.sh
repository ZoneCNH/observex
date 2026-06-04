#!/usr/bin/env bash
set -euo pipefail

tmpdir="$(mktemp -d)"
evidence_out="${DOWNSTREAM_EVIDENCE_OUT:-$tmpdir/downstream-adoption-run.json}"
trap 'rm -rf "$tmpdir"' EXIT

cases=(
  "configx|github.com/ZoneCNH/configx|configx"
  "corekit|example.com/acme/corekit|corekit"
)

for spec in "${cases[@]}"; do
  IFS='|' read -r module_name module_path package_name <<< "$spec"
  out_dir="$tmpdir/$module_name"

  ./scripts/render_template.sh \
    --module-name "$module_name" \
    --module-path "$module_path" \
    --package-name "$package_name" \
    --out "$out_dir"

  ./scripts/check_rendered_template.sh "$out_dir" "$module_name" "$module_path" "$package_name"

  (
    cd "$out_dir"
    git init -q
    git config user.email "ci@example.invalid"
    git config user.name "Template Integration"
    git add .
    git commit -qm "Initial rendered template"

    GOWORK=off go test ./...
    GOWORK=off make contracts
    GOWORK=off make boundary
    downstream_evidence="synthetic downstream smoke: module=${module_path}; template=${module_name}; commands=go test ./..., make contracts, make boundary, make evidence, make release-evidence-check"
    CHECK_STATUS=passed DOWNSTREAM_EVIDENCE="$downstream_evidence" GOWORK=off make evidence
    RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check
  )
done

cat > "$evidence_out" <<JSON
{
  "status": "passed",
  "fixtures": ["configx", "corekit"],
  "command": "GOWORK=off make integration",
  "exit_code": 0,
  "durable_reference": "release/downstream/adoption.json"
}
JSON

echo "downstream evidence: $evidence_out"
echo "integration check passed"
