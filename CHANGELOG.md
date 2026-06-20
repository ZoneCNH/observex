# 变更日志

## v0.3.4 - 2026-06-21

### 修复

- 将当前发布线统一到 `v0.3.4`，同步 package 版本、release manifest fallback、CI 版本推导与发布文档示例。
- 恢复 CI workflow 的版本驱动 release-check 门禁，并继续上传版本化 manifest、sha256 sidecar 和下游证据归档。

### 验证

- `VERSION=v0.3.4 make release-version`
- `GOWORK=off VERSION=v0.3.4 make release-check`

## v0.3.2 - 2026-06-06

### 修复

- 将 release evidence 的版本、manifest 路径、sha256 sidecar 与 tag workflow 绑定到 `v0.3.2` 发布门禁。

### 验证

- 待发布前运行 `VERSION=v0.3.2 make release-final-check`。

## v0.3.1 - 2026-06-06

### 修复

- 将仓库文档中的上游标准来源对齐为 `xlib-standard`，避免继续引用旧标准来源。
- 收紧 renderer 与 rendered-template 检查，确保共享 source token 下 `--module-name` 与 `--package-name` 一致，并阻断旧标准源名称残留。

### 验证

- 已运行 `GOWORK=off make integration`、`GOWORK=off go test ./...`、`GOWORK=off make contracts` 和 `GOWORK=off make boundary`。

## v0.3.0 - 2026-06-04

### 新增

- 明确 observex 是 L1 vendor-neutral 可观测性契约库，而不是通用基础库模板或业务监控模块。
- 新增 L1 ownership ADR 与 evidence 审计页，记录 release manifest、sha256、contract hashes、下游 smoke 和 known gaps 的完成口径。
- 新增 downstream evidence runbook，明确 `external_downstream_unavailable` 时 synthetic smoke 不得宣称为真实生产下游采用。
- 对齐 `HealthReporter` / `ReadinessReporter` / `Attr` root public API、Memory recorder、`latest.json` / sha256 release artifact 生成与校验口径。
- 新增 Public API signature snapshot gate，并把 public Memory 设为 testkit recording 的 canonical model。
- 增强 boundary/import allowlist 和 release evidence 文档标记校验。

### 修复

- `.gitignore` 新增 `.omx/`、`.omc/`、`.worktree/` 规则，阻止 agent 运行时状态入库。
- CI 锁定 `govulncheck@v1.3.0`，确保构建可复现。


## v0.2.0 - 2026-06-01

### 新增

- 新增 `make release-preflight VERSION=vX.Y.Z`，在打 tag 前检查版本、`main` 同步状态、目标 tag、`CHANGELOG.md`、必需工具和最终 release gate。

### 修复

- Release Check workflow 在运行 `make release-check` 前安装 `golangci-lint` 和 `govulncheck`，并使用 `GOWORK=off`，与 CI 的强制 gate 环境保持一致。
- Release Evidence 校验新增目标版本比对，避免目标 tag 与 `manifest.version` 不一致。

## v0.1.0 - 2026-06-01

### 新增

- 初始化 `observex` 结构。
- 添加标准 Go 基础库包骨架。
- 添加 Makefile 命令。
- 添加 Harness Gate 脚本。
- 添加 GitHub Actions 工作流。
- 添加 contracts 文件。
- 添加 Agent 运行时模板。
- 添加 release manifest 模板。
- 添加 typed error、错误包装和 `ErrorKind` contract。
- 添加 client 生命周期、健康检查和请求扩展 metrics contract。
- 添加 health JSON contract 与 contracts 回归测试。
- 添加 config schema 到 `Config` 字段映射的 contract 回归测试。
- 添加 `scripts/render_template.sh`，支持生成 `foundationx` 等具体基础库。
- 添加 `examples/basic`、`examples/config` 和 `examples/health` smoke 测试，锁定文档示例输出。
- 添加 `testkit` 夹具和断言回归测试。
- 添加配置属性测试、配置 fuzz smoke 测试、健康状态 golden 测试和 `testkit` golden 文件工具。

### 安全

- 添加 Secret Gate。
- `make security` 强制运行 `govulncheck ./...` 和密钥扫描；缺少 `govulncheck` 时必须失败。
- 配置脱敏规则覆盖 release Evidence 和日志可见内容。
- Boundary Gate 同时拦截 `github.com/bytechainx/x.go` 和 `github.com/ZoneCNH/x.go`。

### 治理

- 添加 Evidence 和复盘模板。
- CI 在 `make ci` 前安装 `golangci-lint` 和 `govulncheck`，与 Makefile 强制 gate 对齐。
- `make release-check` 统一执行 CI、integration 和 manifest 生成。
- `make release-final-check` 在发布前串联 `release-check`、release Evidence 校验和工作区洁净校验。
- `make integration` 通过临时 `foundationx` 和 `corekit` 渲染、测试、contracts、boundary 与 Evidence 生成验证模板链路。
- `release/manifest/latest.json` 作为生成产物保留在源码历史之外，避免 release Evidence 与源码提交互相污染。

### 验证

- 发布前已运行 `GOWORK=off make release-final-check`。
- `go fmt ./...`、`go vet ./...`、`golangci-lint run ./...`、`go test ./...`、`go test -race ./...`、Boundary、Security、contracts、integration 和 release Evidence 校验均通过。
- `v0.1.0` 为 annotated tag，指向提交 `b6dfe9b93e4417a3b7e077cec1b4c0fffdc37240`。
