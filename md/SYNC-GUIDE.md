# workbuddy2api-panel 同步执行契约

> 跟随基线：`main`（= `upstream/main` 的只读镜像，只做 `merge --ff-only`；起始 2026-10-29）
> 项目自带的分支说明见根目录 `README.local.md`；本文件只管**同步动作与判据**，不重复项目介绍。

## 0. 触发

- 「巡检 / 上游有没有更新 / 读下同步文档」→ 执行 skill §2-A 六步巡检，跑完再汇报。
- 「上游没了 / 分支没了」→ 执行 skill §4 建新基线。

## 1. 分支

| 分支 | 角色 | 规则 |
|---|---|---|
| `local` | 二开活分支，跟踪 `origin/local` | 只向前；**不 rebase、不 merge 上游** |
| `main` | 上游镜像（`BASE_REF`），跟踪 `upstream/main` | 只 `merge --ff-only`，不放假提交；巡检脚本会自动快进它 |

> ⚠️ 仓库里的 `sync-upstream.cmd` 做的是「`local` rebase 到 `main`」——**已弃用**。
> 原因：`local` 已推到 `origin/local`，rebase 会改写已发布历史，必须 force push，多台机器一参与就出事。
> 正确做法见 §3：`cherry-pick -x` 搬提交。

## 2. 能力基线（判收窄的度量尺）

| 能力 | 基线值 | 怎么量 |
|---|---|---|
| 面板路由数 | 65 | `grep -rhoE 'HandleFunc\("' internal/panel \| wc -l` |
| 服务端路由数 | 27 | `grep -rhoE 'HandleFunc\("' internal/server \| wc -l` |
| 模型白名单逻辑 | 存在 `NormalizeAllowedModels` | `grep -rn NormalizeAllowedModels internal/livecfg` |
| 账号池选号链路 | `pick.go` / `cooldown.go` / `plan.go` 均在 | `ls internal/pool` |

数字变少 = 收窄 = 不跟，先问清楚再说。

> 基线变更史：54 / 17（2026-10-29 起）→ **63 / 19**（2026-11-02，随上游 41 条提交跟入：
> 代理池与按账号路由 +9 面板路由、Claude Code 计数端点 +2 服务端路由）
> → **65 / 27**（2026-11-06，随上游 40 条提交跟入：账号分组 / 事件日志清空 +2 面板路由，
> 视频端点 + MCP + 水印裁切 +8 服务端路由）。
>
> ⚠️ 计数口径是**出现次数**（含测试文件里的 httptest mux），不能单看数字判收窄。
> 判「有没有被砍能力」用路由集合差集：
> `comm -23 <(git grep -hoE 'HandleFunc\("[^"]+"' local -- internal/panel internal/server | sort -u) <(git grep -hoE 'HandleFunc\("[^"]+"' main -- internal/panel internal/server | sort -u)`
> 输出为空 = 本地没有上游缺的路由，再用 `comm -13` 看新增。

## 3. 同步动作

```bash
# 1. 巡检（只读）
bash ~/.workbuddy/skills/fork-upstream-sync/scripts/inspect.sh .

# 2. 回滚点
git tag pre-sync-$(date +%Y%m%d-%H%M%S)

# 3. 逐提交搬（用户圈定后才做）
git switch local
git cherry-pick -x <upstream-sha>
#    提交信息前缀改 port:，并剥离收窄 hunk

# 4. 校验
go build ./... && go test ./...
git grep -n misakano7545 -- '*.go' 'go.mod'   # 只允许出现在 module 路径与 import

# 5. 推远端（两个分支都要推，少一个 fork 首页就会显示「上游没更新」）
git push origin local
git push origin main        # 纯镜像 ff，不产生提交
```

## 4. 校验清单（每次必跑）

- [ ] `git status --porcelain` 为空
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过（282 个测试文件，慢是正常的）
- [ ] 路由计数回到 65 / 27
- [ ] `go.mod` 的 module 路径仍是我们 fork 的（若上游改名需评估）
- [ ] `md/SYNC-LOG.md` 已追加一条，含回滚 tag
- [ ] `origin/local` 与 `origin/main` 均已推送（`main` 只允许 ff；用 `git ls-remote` 复核，别看 fork 页面）

> ⚠️ **哪些用例在 Agent 沙箱里跑不了**：`internal/panel` 里 52 条前端用例要 fork `node.exe`，
> `internal/updater` 里 4 条集成用例要 fork 临时编出来的 `test.test.exe`。在 Agent 的 Bash/PowerShell
> 里跑 `go test`，它们会全部报 `fork/exec …: Access is denied`——**这不是回归**（同一份代码在宿主
> 终端里过）。判回归的硬办法：`git worktree add ../_check <上一个 pre-sync-tag>`，在那边跑同一个包，
> 失败条数相同即环境问题。本次实测 panel 50（同步前）→ 52（同步后，新增 2 条 node 用例）。

## 5. 记录去哪看

历史（巡检结论、移植明细、已判定不跟项）全部在 `md/SYNC-LOG.md`。
