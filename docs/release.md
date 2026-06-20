# 发布

## Release Gate

- `make ci`
- `make integration`
- `VERSION=vX.Y.Z make release-version`
- `VERSION=vX.Y.Z make evidence`
- `VERSION=vX.Y.Z make release-evidence-check`

推荐入口是：

```bash
GOWORK=off VERSION=v0.3.4 make release-check
```

`GOWORK=off` 用于证明模板不依赖父级 workspace。

发布前的最终入口是：

```bash
GOWORK=off VERSION=v0.3.4 make release-final-check
```

`release-final-check` 会在完整 gate 之后要求 `release/manifest/v<version>.json` 与当前 HEAD、源码摘要、contract 指纹和依赖清单一致，并要求 git 工作区为 `clean`。它适合在打 tag 或发布前运行；开发中的 `release-check` 允许工作区因为未提交改动显示为 `dirty`，但仍会校验 manifest 与当前内容一致。

打 tag 前推荐使用 release preflight：

```bash
make release-preflight VERSION=v0.3.4
```

`release-preflight` 会先检查版本号、当前分支、工作区洁净状态、`main` 与 `origin/main` 是否一致、目标 tag 是否已存在、`CHANGELOG.md` 是否包含目标版本，以及 `golangci-lint` / `govulncheck` 是否已安装；随后以 `GOWORK=off` 运行 `release-final-check`。tag 应在该入口通过后再创建和推送。

## Required Release Check

`make release-check` 是默认发布门禁，必须通过：

```text
ci
integration
evidence
release-evidence-check
```

## Extended Release Check

`make release-check-extended` 是发布前强验证，推荐在重要版本、公共 API 变更、contract 变更、schema 变更、metrics 变更时执行：

```text
ci-extended
integration
evidence
release-evidence-check
```

`make ci-extended` 会在默认 `ci` 外追加：

```text
property
golden
fuzz-smoke
```

## Gate 工具契约

`make ci` 中的 `make lint` 和 `make security` 是强制 gate。运行前必须可用：

- `golangci-lint`
- `govulncheck`

缺少任一工具时，本地 Makefile 必须硬失败。GitHub Actions CI 和 Release Check workflow 会在运行 `make ci` / `make release-check` 前安装 `golangci-lint` 和 `govulncheck`，以保证本地与远端 workflow 对同一组强制 gate 负责。

`make security` 必须同时运行 `govulncheck ./...` 和 `scripts/check_secrets.sh`；不得把漏洞扫描降级为可选检查。

## Evidence

发布 Evidence 必须显式传入 `VERSION=vX.Y.Z`；`make release-version` 会在生成或校验 Evidence 前确认该值与 `pkg/observex/version.go` 一致。默认生成路径由 `VERSION` 决定，例如 `release/manifest/v0.3.4.json`，也可通过 `RELEASE_MANIFEST=...` 覆盖路径。release workflow 还必须发布版本化 manifest、版本化 sha256 sidecar、`release/manifest/latest.json` 和 `release/manifest/latest.json.sha256`，且 `latest.json` 必须与版本化 manifest 字节一致，便于下游和审计流程以稳定路径读取最新 manifest。manifest 文件是生成产物，不提交到源码历史；提交到仓库的是 `release/manifest/template.json`；CI release workflow 会上传 `release/manifest/*.json` 和 sha256 文件作为 artifact。

版本化 manifest 至少包含：

- `module`
- `version`
- `commit`
- `tree_sha`
- `source_digest`
- `tracked_file_count`
- `go_version`
- `generated_at`
- `generated_by`
- `tree_state`
- `checks`
- `contracts`
- `dependencies`
- `tools`
- `artifacts`
- `downstream_adoption`
- `notes`

`make release-check` 成功后会以 `CHECK_STATUS=passed` 和同一个 `VERSION` 生成 manifest，并立即以同一个 `VERSION` 运行 `make release-evidence-check`；该校验会先执行 `scripts/check_downstream_evidence.sh`，确保 `release/downstream/adoption.json` 以 `fixture_smoke` 和 `real_adoption` 分离记录合成 downstream fixtures、执行命令和真实下游 blocker。若单独运行 `make evidence`，未显式传入的检查状态默认为 `unknown`，后续校验会拒绝把这些状态当作已通过的 release gate。因为版本化 manifest 不再提交，manifest 中的 `commit` 可以指向实际执行 release gate 的 HEAD，避免自引用提交哈希导致的永久漂移。

完成声明必须额外记录 provider dependency scan、examples smoke、contract hashes、manifest hash、命令退出码、下游 smoke 或精确 blocker。Extended Evidence 推荐额外记录：

- `make ci-extended` 结果。
- `make property` 结果。
- `make fuzz-smoke` 结果。
- `make golden` 结果。
- compatibility 和 observability contract 结果。

`source_digest` 基于 `git ls-files` 中的受跟踪文件内容计算；`contracts` 固定记录核心 contract 文件（包含 `contracts/public_api.snapshot`）的 SHA256；`dependencies` 来自 `go list -m -json all`；`tools` 记录 Go、`golangci-lint` 和 `govulncheck` 的版本或可用状态；`downstream_adoption` 记录 `fixture_smoke` 下游验证状态/命令退出码，以及独立的 `real_adoption` consumers 或真实下游 blocker。这些字段由 `internal/tools/releasemanifest` 生成并校验，不再由 shell 拼接 JSON。

`make integration` 会调用 `scripts/render_template.sh` 生成临时 `configx` 和 `corekit` 两个下游库，并对每个生成目录执行：

- 模块路径、包目录和旧模板标识扫描。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed DOWNSTREAM_EVIDENCE=<synthetic downstream smoke evidence> VERSION=<rendered version> GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 VERSION=<rendered version> GOWORK=off make release-evidence-check`

这一步用于证明模板替换、包目录迁移、imports、contracts、边界检查和生成后 release Evidence 都能在下游库中独立工作。脚本还会输出一次运行级 downstream evidence JSON；源码历史中的 durable reference 是 `release/downstream/adoption.json`。

## 规则

- 没有 Evidence artifact、manifest sha256、contract hash 记录和 downstream adoption/blocker 记录不得发布。
- `tree_state` 为 `dirty` 时可以在开发中生成 Evidence，但正式发布前必须通过 `make release-final-check`。
- 不得在 release manifest、PR、Issue 或变更日志条目中包含原始凭据。
- 不得依赖 `github.com/ZoneCNH/x.go`。
- public API、config schema、error kind、health JSON、metrics name、downstream adoption schema 或 release manifest schema 变更必须在 release notes 或 release manifest 中显式标记 breaking change。


## Downstream Adoption Evidence

`release/downstream/adoption.json` 是提交到源码历史的 durable downstream 证据索引。当前证据是 fixture-backed：`configx` 和 `corekit` 由 `scripts/run_integration.sh` 渲染并完整运行测试、contracts、boundary、evidence 与 release evidence check。由于本仓库不包含维护中的真实外部下游仓库，`external_real_downstream` blocker 必须保留，直到替换为真实仓库 commit、版本和 CI 证据。

正式发布声明必须区分：fixture-backed downstream smoke 已通过；真实外部下游采用仍需单独证据或继续声明 blocker。
