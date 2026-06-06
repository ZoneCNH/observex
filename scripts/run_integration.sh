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
  "fixture_smoke": {
    "status": "passed",
    "fixtures": ["configx", "corekit"],
    "commands": [
      {
        "command": "GOWORK=off make integration",
        "status": "passed",
        "exit_code": 0,
        "evidence": "scripts/run_integration.sh"
      }
    ]
  },
  "real_adoption": {
    "status": "blocked",
    "consumers": [],
    "blockers": [
      {
        "scope": "external_real_downstream",
        "reason": "Synthetic fixture smoke completed; real external downstream adoption remains represented by the durable release/downstream/adoption.json blocker.",
        "evidence": "release/downstream/adoption.json"
      }
    ]
  },
  "durable_reference": "release/downstream/adoption.json"
}
JSON

echo "downstream evidence: $evidence_out"
echo "integration check passed"
