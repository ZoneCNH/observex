# 生成模板

## 用途

`scripts/render_template.sh` 用于把 `observex` 渲染为具体下游基础库，例如 `configx`。脚本负责同步替换 module path、共享 source token、`pkg/` 目录名、imports、文档占位符和脚本中的模板名称。

当前模板中 `observex` 同时承担 source module name 和 package name，因此 `--module-name` 必须与 `--package-name` 一致。若未来需要把二者分离，应先引入显式占位符，再调整渲染规则，避免裸 token 被错误替换。

`foundationx` 是本模板的上游基础依赖，不作为 `observex` 的渲染目标；否则生成后的 `github.com/ZoneCNH/foundationx` 会把上游依赖路径解析为自身包，形成 Go import cycle。

## 标准基线

当前生成模板对齐 `xlib-standard` `v0.4.19`（`4463a60`）。该基线新增的 L2 adapter contract pack、`templates/l2`、`scripts/verify_l2_standard.py` 和 L2 readiness evidence 属于标准源能力，用于后续 L2 adapter 仓库与 `testkitx`、`xlibgate` 协作。

`observex` 当前仍是 L1 可观测性契约库，生成脚本只承担 L1 基础库渲染、boundary、contract、release Evidence 和下游 smoke 验证。同步 `xlib-standard` 时不得把 L2 adapter 模板或 provider gate 复制到本仓库，除非目标明确升级为 L2 adapter 标准库。

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

- `github.com/ZoneCNH/observex` 先替换为 `--module-path`，避免 bare token 提前改写 module path。
- `observex` 作为共享 source token，替换为 `--module-name` / `--package-name` 指向的同一名称。
- `pkg/observex` 会移动为 `pkg/<package>`，Go imports 随文本替换同步更新。
- 文档、Go 代码、JSON contract、shell 脚本、Makefile 和 CI 配置同步更新。

脚本不会复制 `.git`、`.omx`、`.omc`、`.worktree` 和 `release/manifest/v*.json`。版本化 manifest 是生成产物，生成后的库必须自己运行 release gate 生成新的 Evidence artifact。

## 验证

生成后至少运行：

```bash
GOWORK=off VERSION=v0.3.4 make release-check
```

模板自身的 `make integration` 会渲染两个临时下游库：

- `configx`：目标仓库路径 `github.com/ZoneCNH/configx`，用于证明 ZoneCNH 下游基础库形态仍可生成。
- `corekit`：中性路径 `example.com/acme/corekit`，用于证明替换逻辑不依赖特定组织或包名。

每个临时库都会运行以下验证：

- `scripts/check_rendered_template.sh`：确认 `go.mod` module path、`pkg/<package>` 目录、旧模板目录、旧 module path、占位符、旧标准源名称和 `observex` 标识。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed VERSION=v0.3.4 GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 VERSION=v0.3.4 GOWORK=off make release-evidence-check`

这组验证用于防止生成脚本、包路径、imports、contract gate、boundary gate 和生成后 Evidence 回归。

## 生成后 Release Evidence

生成后的库会继承 `internal/tools/releasemanifest`。`VERSION=vX.Y.Z` 是 release evidence 入口的必需参数；默认 artifact 路径为 `release/manifest/<VERSION>.json`，可通过 `RELEASE_MANIFEST=...` 覆盖，但版本化路径必须与 manifest version 一致。manifest 包括当前 HEAD、tree SHA、源码摘要、contract SHA256、依赖清单和工具版本。发布前应使用：

```bash
GOWORK=off VERSION=v0.3.4 make release-final-check
```

`release-final-check` 要求所有 gate 状态为 `passed`，并要求 git 工作区为 `clean`。如果只是开发中自测，`VERSION=v0.3.4 make release-check` 已足够；它允许工作区显示 `dirty`，但仍会验证 manifest 和当前源码内容一致。

## 边界

生成后的基础库仍必须保持独立，不能依赖 `github.com/ZoneCNH/x.go` 或任何 `x.go/internal/*` 包。上层 `x.go` 可以依赖生成后的基础库，但依赖方向不能反转。
