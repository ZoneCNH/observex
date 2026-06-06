.PHONY: fmt
fmt:
	GOWORK=off go fmt ./...

.PHONY: vet
vet:
	GOWORK=off go vet ./...

.PHONY: test
test:
	GOWORK=off go test ./...

.PHONY: race
race:
	GOWORK=off go test -race ./...

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		GOWORK=off golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed"; \
		exit 1; \
	fi

.PHONY: integration
integration:
	./scripts/run_integration.sh

.PHONY: security
security:
	@if command -v govulncheck >/dev/null 2>&1; then \
		GOWORK=off govulncheck ./...; \
	else \
		echo "govulncheck not installed"; \
		exit 1; \
	fi
	./scripts/check_secrets.sh

.PHONY: boundary
boundary:
	./scripts/check_boundary.sh

.PHONY: contracts
contracts:
	./scripts/check_contracts.sh

.PHONY: property
property:
	GOWORK=off go test ./... -run 'Test.*Property|Test.*Invariant'

.PHONY: fuzz-smoke
fuzz-smoke:
	./scripts/run_fuzz_smoke.sh

.PHONY: golden
golden:
	GOWORK=off go test ./... -run 'Test.*Golden|Test.*Snapshot'

.PHONY: examples
examples:
	GOWORK=off go test ./examples/...

.PHONY: release-version
release-version:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required"; exit 1; fi
	@release_version="$(VERSION)"; \
	printf '%s\n' "$$release_version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?$$' || { echo "VERSION must look like vX.Y.Z, got: $$release_version"; exit 1; }; \
	package_version="$$(sed -nE 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"([^"]+)".*/\1/p' pkg/observex/version.go | head -n1)"; \
	if [ -z "$$package_version" ]; then echo "could not determine package version from pkg/observex/version.go"; exit 1; fi; \
	if [ "$$release_version" != "$$package_version" ]; then echo "VERSION $$release_version does not match pkg/observex/version.go ($$package_version)"; exit 1; fi

.PHONY: evidence
evidence: release-version
	./scripts/generate_manifest.sh

.PHONY: release-evidence-check
release-evidence-check: release-version
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 VERSION="$(VERSION)" ./scripts/check_release_evidence.sh

.PHONY: ci
ci: fmt vet lint test race examples boundary security contracts

.PHONY: ci-extended
ci-extended: ci property golden fuzz-smoke

.PHONY: release-check
release-check: release-version ci integration
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required for release-check"; exit 1; fi
	CHECK_STATUS=passed VERSION="$(VERSION)" $(MAKE) evidence
	VERSION="$(VERSION)" $(MAKE) release-evidence-check

.PHONY: release-check-extended
release-check-extended: release-version ci-extended integration
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required for release-check-extended"; exit 1; fi
	CHECK_STATUS=passed VERSION="$(VERSION)" $(MAKE) evidence
	VERSION="$(VERSION)" $(MAKE) release-evidence-check

.PHONY: release-final-check
release-final-check: release-version
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required for release-final-check"; exit 1; fi
	GOWORK=off VERSION="$(VERSION)" $(MAKE) release-check
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 RELEASE_EVIDENCE_REQUIRE_CLEAN=1 VERSION="$(VERSION)" ./scripts/check_release_evidence.sh

.PHONY: release-preflight
release-preflight:
	./scripts/check_release_preflight.sh "$(VERSION)"
	GOWORK=off VERSION="$(VERSION)" $(MAKE) release-final-check
