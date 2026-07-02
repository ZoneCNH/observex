# 仓库贡献指南

## 项目结构与模块组织

本仓库是 Go 1.23 基础库模板，模块路径为 `github.com/ZoneCNH/observex`。公共 API 位于 `pkg/observex`；仅供内部使用的实现放在 `internal/`，当前包括 `sanitize`、`validation` 和 `runtime`。可复用测试工具在 `testkit/`，示例程序在 `examples/basic`、`examples/config` 和 `examples/health`。contracts 文件位于 `contracts/`，项目说明在 `docs/`，自动化脚本在 `scripts/`，发布 Evidence 生成到 `release/manifest/`。

## 构建、测试与开发命令

- `make fmt`：执行 `go fmt ./...`。
- `make vet`：执行 `go vet ./...`。
- `make test`：运行全部单元测试。
- `make race`：使用 race detector 运行测试。
- `make lint`：执行 `golangci-lint run ./...`；缺少 `golangci-lint` 时必须失败。
- `make security`：执行 `govulncheck ./...` 和 `scripts/check_secrets.sh`；缺少 `govulncheck` 时必须失败。
- `make ci`：运行格式化、vet、lint、测试、race、Boundary、Security 和 contracts 检查。
- `make release-check`：运行 CI、集成测试和 Evidence 生成。
- `make evidence`：生成 release manifest。

## 编码风格与命名约定

使用标准 Go 风格：交给 `gofmt` 处理缩进，包名保持简短，导出标识符要清晰表达用途。模板占位符如 `observex`、`github.com/ZoneCNH/observex`、`observex` 必须在代码和文档中保持一致。公共库能力放入 `pkg/observex`，私有辅助逻辑放入 `internal/`。不要新增隐藏全局客户端，不要隐式读取生产密钥，不要引入业务领域模型。

## 测试规范

测试使用 Go 标准 `testing` 包，命名遵循 `TestXxx`；场景较多时优先使用表驱动测试。必须覆盖配置校验、配置脱敏、客户端创建、幂等关闭、健康和非健康状态检查、示例 smoke 输出、`testkit` 夹具和 contracts 映射。小改动至少运行 `go test ./...`；涉及并发时运行 `make race`；影响发布流程时运行 `make integration`。

## 提交与 Pull Request 规范

提交信息必须遵循 Lore protocol：第一行说明变更意图，正文使用 `Constraint:`、`Rejected:`、`Confidence:`、`Scope-risk:`、`Directive:`、`Tested:` 和 `Not-tested:` 等 trailer 记录决策和验证。PR 需要说明对模板或库行为的影响，关联相关 issue，列出已运行命令，并说明生成的 Evidence artifact，例如 `release/manifest/latest.json`。`latest.json` 是生成产物，不提交到源码历史；只有文档渲染或界面变化需要截图。

## 安全与边界规则

基础库必须独立于 `x.go`，且不能包含业务专用模型。依赖、包结构或命名发生变化后运行 `make boundary`；提交 PR 前运行 `make security`。不要提交真实凭据，`scripts/check_secrets.sh` 会扫描常见密钥模式。

## 文档语言规则

所有仓库文档必须默认使用中文叙述，包括 `README.md`、`docs/`、`.agent/`、`contracts/*.md`、变更日志、发布说明、PR 描述模板和贡献指南。专业术语、代码标识符、命令、路径、包名、外部专有名词、协议固定短语和提交标题必须保留项目惯用原文，例如 Agent、Harness、manifest、schema、CI、PR、Issue、Go module。新增或更新文档时，先检查是否存在整段英文说明；除非用户明确要求英文，否则应改写为中文，但不要翻译专业术语。

## Agent 专用说明

仓库协作、代码评审和进度更新默认使用中文，除非用户明确要求其他语言。代码、命令、路径、包名和提交标题保留项目惯用语言。
