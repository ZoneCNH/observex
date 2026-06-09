# 下游采用与 blocker Evidence

日期：2026-06-04

本页是 observex release evidence 的持久下游记录。它区分“合成 downstream smoke”和“真实持久下游采用”，避免把本地模板渲染误读为生产级 adoption。

## 当前结论

- 合成 downstream smoke：由 `GOWORK=off make integration` 驱动，生成临时 `configx` 与 `corekit` 模块并独立运行 test、contracts、boundary、evidence 和 release evidence check。
- 真实持久下游采用：当前本地仓库没有可验证的外部消费仓库、commit、CI 日志或版本升级记录。
- blocker 标记：`external_real_downstream`。
- 发布口径：在补齐真实下游证据前，observex 可以声明本地 L1 结构 gate 与模板化 downstream smoke 通过；不得声明已经完成生产级真实下游采用。


## 机器可校验 source record

`release/downstream/adoption.json` 必须用两个顶层对象分离证据语义：

- `fixture_smoke`：只记录合成 fixture smoke。必须包含 `status`、`fixtures` 和 `commands`，其中 fixtures 至少包括 `configx` 与 `corekit`，命令必须记录 `GOWORK=off make integration`、`status`、`exit_code` 和 evidence 来源。
- `real_adoption`：只记录真实外部 consumer 采用。`status=passed` 时必须包含 consumer、repository、commit、observex version、命令和 evidence；真实 consumer 不可用时必须保留 `status=blocked`、空 `consumers` 和 `external_real_downstream` blocker。

禁止在顶层重新引入旧的 `status` / `fixtures` / `commands` / `blockers` 字段；否则合成 smoke 会再次和真实采用混淆。

## 真实下游 evidence 最低字段

| 字段 | 要求 |
|---|---|
| Consumer | 下游库或服务名称 |
| Repository | 仓库 URL 或本地路径 |
| Commit | 下游 commit SHA；不得只写分支名 |
| observex version | module version、replace 指向或 commit SHA |
| Commands | 实际运行命令，例如 `GOWORK=off go test ./...`、contracts、boundary、release evidence |
| Result | 退出码、关键日志、manifest/hash 路径 |
| Failure/blocker | 若失败，记录精确错误和下一步；若不可用，使用 `external_real_downstream` |
| Conclusion | adoption passed / blocked / non-final |

## 本地合成 smoke 记录

| Evidence | 命令 | 结论 |
|---|---|---|
| Template downstream | `GOWORK=off make integration` | 证明 `configx` 与 `corekit` 临时模块可被渲染、测试并通过 contracts/boundary/evidence；这是模板压力测试，不是外部真实 adoption。 |

## Release 使用规则

1. `docs/evidence.md` 和 release manifest 可以引用本页作为 downstream evidence source。
2. 当真实下游不可用时，release evidence 必须保留 `external_real_downstream` 并把结论标记为 non-final blocker。
3. 只有记录了真实 consumer、repo/commit、observex version、命令和通过结果后，才允许把该记录升级为真实下游 adoption evidence。
