# 生成模板

## 用途

`scripts/render_template.sh` 用于把 `observex` 渲染为具体下游基础库，例如 `configx`。脚本负责同步替换 module name、module path、package name、`pkg/` 目录名、imports、文档占位符和脚本中的模板名称。

`foundationx` 是本模板的上游基础依赖，不作为 `observex` 的渲染目标；否则生成后的 `github.com/ZoneCNH/foundationx` 会把上游依赖路径解析为自身包，形成 Go import cycle。

## 示例

```bash
scripts/render_template.sh \
  --module-name configx \
  --module-path github.com/ZoneCNH/configx \
  --package-name configx \
  --out ../configx
```

`--out` 必须指向不存在或为空的目录，避免覆盖已有仓库内容。

## 渲染范围

- `observex` 替换为 `--module-name`。
- `github.com/ZoneCNH/observex` 和 `github.com/ZoneCNH/observex` 替换为 `--module-path`。
- `observex`、`pkg/observex` 和 `observex` imports 替换为 `--package-name`。
- 文档、Go 代码、JSON contract、shell 脚本、Makefile 和 CI 配置同步更新。

脚本不会复制 `.git`、`.omx`、`.worktree` 和 `release/manifest/v*.json`。版本化 manifest 是生成产物，生成后的库必须自己运行 release gate 生成新的 Evidence artifact。

## 验证

生成后至少运行：

```bash
GOWORK=off make release-check
```

模板自身的 `make integration` 会渲染两个临时下游库：

- `configx`：目标仓库路径 `github.com/ZoneCNH/configx`，用于证明 ZoneCNH 下游基础库形态仍可生成。
- `corekit`：中性路径 `example.com/acme/corekit`，用于证明替换逻辑不依赖特定组织或包名。

每个临时库都会运行以下验证：

- `scripts/check_rendered_template.sh`：确认 `go.mod` module path、`pkg/<package>` 目录、旧模板目录、旧 module path、占位符和 `observex` 标识。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`

这组验证用于防止生成脚本、包路径、imports、contract gate、boundary gate 和生成后 Evidence 回归。

## 生成后 Release Evidence

生成后的库会继承 `internal/tools/releasemanifest`。该工具默认生成并校验 `release/manifest/v0.3.0.json`，可通过 `VERSION=vX.Y.Z` 或 `RELEASE_MANIFEST=...` 指定路径，其中包括当前 HEAD、tree SHA、源码摘要、contract SHA256、依赖清单和工具版本。发布前应使用：

```bash
GOWORK=off make release-final-check
```

`release-final-check` 要求所有 gate 状态为 `passed`，并要求 git 工作区为 `clean`。如果只是开发中自测，`make release-check` 已足够；它允许工作区显示 `dirty`，但仍会验证 manifest 和当前源码内容一致。

## 边界

生成后的基础库仍必须保持独立，不能依赖 `github.com/ZoneCNH/x.go` 或任何 `x.go/internal/*` 包。上层 `x.go` 可以依赖生成后的基础库，但依赖方向不能反转。
