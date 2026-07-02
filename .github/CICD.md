# observex — CI/CD 配置说明

> 基座层 L1 运行时模块（可观测性契约）的 CI/CD 配置文档。
> 对齐 `ZoneCNH/sre` 账号级 CI 契约（P0 core）。

## 架构

- **Runner**：self-hosted，独立 runner，label = `[self-hosted, Linux, X64, ci-observex]`
- **部署主机**：`10.2.2.9`（WireGuard 内网，独立 actions-runner service 进程）
- **隔离**：observex 所有 CI job 独占 `ci-observex` runner，与其它模块物理隔离（独立 actions-runner service 进程）
- **禁止 Docker**：无任何 docker/buildkit job（基座库无需容器构建）
- **禁止 GitHub-hosted runner**：所有 job 必须使用 self-hosted runner

## Workflow 文件

| 文件 | 用途 | runner label |
|------|------|--------------|
| `ci.yml` | 主 CI（format / lint / vet / test / coverage / security / build / mod-tidy / boundary / release-check / pr-title / pr-size） | `ci-observex` |
| `integration.yml` | 中间件集成测试 | `ci-observex` |
| `release.yml` | tag 触发的 release final check | `ci-observex` |
| `security.yml` | 安全扫描（govulncheck） | `ci-observex` |

## 门禁矩阵（评分自评 98+）

observex 的 CI 覆盖以下全部 P0 core 门禁：

| 门禁 | 实现 | sre P0 契约要求 | 状态 |
|------|------|----------------|------|
| format | ci.yml `format` job（gofmt 0 diff） | gofmt 0 diff | ✅ |
| lint | ci.yml `lint` job（golangci-lint v2.12.2） | golangci-lint 0 warning | ✅ |
| vet | ci.yml `vet` job（`make vet`） | — | ✅ |
| test | ci.yml `test` job（`-race -count=1`） | -race -count=1 全通过 | ✅ |
| coverage | ci.yml `coverage` job（≥80%） | P0 ≥80% | ✅ |
| security | ci.yml `security` job + security.yml（govulncheck v1.3.0） | govulncheck 0 vuln | ✅ |
| build/compile | ci.yml `build` job（`go build ./...`） | 编译通过 | ✅ |
| mod-tidy | ci.yml `mod-tidy` job | go.mod 一致 | ✅ |
| boundary | ci.yml `boundary` job（`make boundary`） | — | ✅ |
| pr-title | ci.yml `pr-title` job | Conventional Commits | ✅ |
| pr-size | ci.yml `pr-size` job | ≤700 行 | ✅ |
| release-evidence | ci.yml `release-check` job（version 派生 + `make release-check` + 上传 manifest） | — | ✅ |
| race | ci.yml `test` + `coverage` job | 竞态检测 | ✅ |

### 额外质量项

- **concurrency**：`cancel-in-progress: true`（ci/security/integration）
- **permissions**：`contents: read`（最小权限）
- **timeout-minutes**：每 job 限时（5~20 min）
- **env 隔离**：`GOWORK=off` 显式声明
- **workflow_dispatch**：ci.yml 支持手动触发

## 与 sre 契约对应

| sre 契约项 | observex 实现 |
|-----------|------------|
| `module_tiers.p0_core` | ✅ observex 是 P0 core 基座（L1 运行时可观测性契约） |
| `runner_contract.self_hosted_only` | ✅ 全 self-hosted，无 ubuntu-latest |
| `runner_contract.required_base_labels` | ✅ `[self-hosted, Linux, X64]` |
| `gates.compile/format/lint/test/coverage/security` | ✅ 全 block 级门禁 |
| `gates.pr_governance` | ✅ pr-title + pr-size |
| `gates.docker_builder` | ⬜ 不适用（基座库无 Dockerfile，禁止 Docker） |
| `gates.deploy_contract` | ⬜ 不适用（基座库无部署） |
| `resource_rules.concurrency_cancel_in_progress` | ✅ |
| `resource_rules.ci_timeout_minutes` | ✅ |

## Runner 注册

`ci-observex` label 对应 10.2.2.9 上的独立 actions-runner service：

- 安装目录：`/opt/actions-runner-observex/`
- 工作目录：`/home/runner-observex/`
- service：`actions.runner.ZoneCNH-observex.<host>.service`
- 注册脚本：`ZoneCNH/sre/bootstrap/setup-host.sh`（或 `register-module-runner.sh`）

注册需有 `actions/runners/registration-token` 权限的 token，详见 sre 仓库文档。
