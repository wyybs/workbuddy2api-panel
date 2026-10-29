# workbuddy2api-panel 同步日志（append-only，只加不改）

## 已判定不跟（提案前先看这里）

| 上游提交 / 范围 | 内容 | 为什么不跟 |
|---|---|---|

## 时间线（新→旧）

### 2026-10-29 15:18 基线改指 main（不再用 upstream/main 比较）

- `BASE_REF` 由 `upstream/main` 改为 `main` —— 本仓库既有约定是 main 只做 ff 等于 upstream/main，
  没必要在比较时绕过它。
- `inspect.sh` 已支持「基线是本地镜像分支时自动快进」：本次实测 `main` bb6d581 -> 9232c89，未产生提交。
- 说明：`upstream/main` 是 remote-tracking 引用（fetch 自动生成），不是我们建的分支；本地分支始终只有 main / local。

### 2026-10-29 15:05 接入 fork-upstream-sync（首次基线）

- 基线：`local` @ 235269a / `upstream/main` @ bb6d581（分叉点 bb6d581，**local 领先 0 / 落后 2**）
- `local` 独有 2 个提交，全是文档与脚本：
  - `802e532` chore: 增加 sync-upstream.cmd（拉上游 main 并 rebase local）
  - `235269a` docs: 增加 README.local.md（分支模型/构建/部署/排程/踩坑）
- 能力基线：面板路由 54 / 服务端路由 17 / `NormalizeAllowedModels` 在
- 基线 tag：`base-20261029`
- 备注：`sync-upstream.cmd` 的 rebase 策略已弃用（`local` 已推远端，rebase 需 force push），
  后续同步改用 `cherry-pick -x`
