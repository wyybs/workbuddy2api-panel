# workbuddy2api-panel 同步日志（append-only，只加不改）

## 已判定不跟（提案前先看这里）

| 上游提交 / 范围 | 内容 | 为什么不跟 |
|---|---|---|

## 时间线（新→旧）

### 2026-10-29 15:26 推送 fork + 部署到 192.168.2.22

- 推送：`origin/local` `235269a → db41877`（含 2 个 port 提交、Dockerfile 修复、文档与日志）。
- 服务器仓库原先的坑：`/home/appsrv/workbuddy2api-panel` 的 `origin` 指的是**上游**仓库，
  且跑在 `main` @ bb6d581（4 小时前构建，没有本次两个修复）。新增远端 `fork` 指向本 fork 后切到 `local`。
- **Dockerfile 的 GOPROXY 三行此前只存在于服务器工作区**（`Dockerfile.orig.bak` 旁证），从未入库；
  README 4.2 却写着「已内置」。本次补提交 `f4c6d3e` 入库，并在服务器 `git checkout -- Dockerfile`
  丢掉那份未纳管的手改，避免下次 pull 冲突。
- 部署步骤（已写进 `README.local.md` §5「更新到服务器」）：
  `docker tag …:pre-20261029-1525`（旧镜像留作回滚点）→ `git fetch fork local && git merge --ff-only fork/local`
  → `docker compose up -d --build`。
- 结果：HEAD `db41877`，镜像重建完成，容器 `Up (healthy)`；
  `/healthz` → `{"healthy":2,"realm_servable":{"cn":true,"global":false},"total":2}`；
  `curl /panel/app.js | grep -c openRemarks` = 3（行内「备注」入口已上线）；
  日志无异常，恢复 6 个用量桶 / 29 条积分流水。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261029-1525 workbuddy2api-panel-wb2api:latest` +
  `docker compose up -d --force-recreate`；或 `git checkout c605270 && docker compose up -d --build`。

### 2026-10-29 15:20 移植上游 2 个修复（updater exe 重试 / 账号池行内备注）

- 巡检：`upstream/main` @ 9232c89，`local` 领先 4 / 落后 2，工作区干净，上游可达（8 heads）。
- 回滚 tag：**`pre-sync-20261029-151858`**（= `86af9ac`，`git reset --hard pre-sync-20261029-151858` 可整段退回）。
- 搬入（`cherry-pick -n` + 重写提交信息，保留 `(cherry picked from commit …)` 溯源）：
  - `f43d867` port: fix(updater): 删刚跑过的 exe 要重试 ← `a25192e`
  - `156c618` port: fix(panel): 账号池每行补「备注」按钮 ← `9232c89`
- 收窄判定：**零收窄**。两提交只动 `internal/updater/{updater.go,integration_test.go,remove_test.go}` 与
  `internal/panel/{app.js,index.html,frontend_test.go}`，无 HandleFunc 增删、无白名单逻辑改动，hunk 原样保留。
- 校验：`go build ./...` 通过；`go test ./...` 26 个包全过（1m02s）；
  面板路由 54 / 服务端路由 17 / `NormalizeAllowedModels` 12（与上游同值）；
  `git grep misakano7545` 仅命中 module 路径与 import。
- 结果：`git diff upstream/main..local -- '*.go' '*.js' '*.html'` 为空 —— **代码层与上游完全对齐**；
  剩余差异 8 个文件全是本地独有（`.fork-sync.*`、`md/SYNC-*.md`、`README.local.md`、`README.md` 顶部标注、
  `.gitignore` 放行、`sync-upstream.cmd`）。
- 备注：cherry-pick 后 `local..upstream/main` 仍显示「落后 2」是正常的——搬的是内容不是 SHA，
  判"跟上了"要看上面的代码层 diff 是否为空，不看 rev-list 计数。
- 未推远端：`origin/local` 尚未包含这 2 个 port 提交，需要时 `git push origin local`。

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
