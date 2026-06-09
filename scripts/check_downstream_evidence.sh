#!/usr/bin/env bash
set -euo pipefail

evidence="${DOWNSTREAM_EVIDENCE:-release/downstream/adoption.json}"

if [[ ! -f "$evidence" ]]; then
  echo "ERROR: missing downstream evidence: $evidence" >&2
  exit 1
fi

python3 - "$evidence" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
try:
    data = json.loads(path.read_text())
except json.JSONDecodeError as exc:
    print(f"ERROR: downstream evidence is not valid JSON: {exc}", file=sys.stderr)
    sys.exit(1)

failures = []

def require(condition, message):
    if not condition:
        failures.append(message)

legacy = {"status", "fixtures", "commands", "blockers"} & set(data)
require(not legacy, f"legacy top-level downstream fields are not allowed: {', '.join(sorted(legacy))}")

fixture = data.get("fixture_smoke")
real = data.get("real_adoption")
require(isinstance(fixture, dict), "fixture_smoke object is required")
require(isinstance(real, dict), "real_adoption object is required")

if isinstance(fixture, dict):
    require(fixture.get("status") == "passed", f"fixture_smoke.status must be passed, got {fixture.get('status')!r}")
    fixtures = fixture.get("fixtures")
    require(isinstance(fixtures, list) and fixtures, "fixture_smoke.fixtures is required")
    fixture_names = {item.get("name") for item in fixtures if isinstance(item, dict)} if isinstance(fixtures, list) else set()
    require({"configx", "corekit"} <= fixture_names, "fixture_smoke.fixtures must include configx and corekit")
    commands = fixture.get("commands")
    require(isinstance(commands, list) and commands, "fixture_smoke.commands is required")
    has_integration = False
    if isinstance(commands, list):
        for index, command in enumerate(commands):
            if not isinstance(command, dict):
                failures.append(f"fixture_smoke.commands[{index}] must be an object")
                continue
            if command.get("command") == "GOWORK=off make integration":
                has_integration = True
            require(command.get("status") == "passed", f"fixture_smoke.commands[{index}].status must be passed")
            require(command.get("exit_code") == 0, f"fixture_smoke.commands[{index}].exit_code must be 0")
            require(bool(command.get("evidence")), f"fixture_smoke.commands[{index}].evidence is required")
    require(has_integration, "fixture_smoke.commands must include GOWORK=off make integration")

if isinstance(real, dict):
    status = real.get("status")
    require(isinstance(status, str) and status, "real_adoption.status is required")
    consumers = real.get("consumers", [])
    blockers = real.get("blockers", [])
    require(isinstance(consumers, list), "real_adoption.consumers must be a list")
    require(isinstance(blockers, list), "real_adoption.blockers must be a list")
    if status == "passed":
        require(bool(consumers), "real_adoption.consumers is required when real adoption is passed")
        for index, consumer in enumerate(consumers if isinstance(consumers, list) else []):
            if not isinstance(consumer, dict):
                failures.append(f"real_adoption.consumers[{index}] must be an object")
                continue
            for field in ("name", "repository", "commit", "observex_version", "commands", "evidence"):
                require(bool(consumer.get(field)), f"real_adoption.consumers[{index}].{field} is required")
    else:
        require(bool(blockers), "real_adoption.blockers is required when real adoption is not passed")
        blocker_scopes = {item.get("scope") for item in blockers if isinstance(item, dict)} if isinstance(blockers, list) else set()
        require("external_real_downstream" in blocker_scopes, "real_adoption.blockers must include external_real_downstream")

if failures:
    for failure in failures:
        print(f"ERROR: {failure}", file=sys.stderr)
    sys.exit(1)
PY

echo "downstream evidence check passed"
