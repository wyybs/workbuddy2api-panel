# workbuddy2api-panel 同步执行契约

> 跟随基线：`main`（= `upstream/main` 的只读镜像，只做 `merge --ff-only`；起始 2026-10-29）
> 项目自带的分支说明见根目录 `README.local.md`；本文件只管**同步动作与判据**，不重复项目介绍。

## 0. 触发

- 「巡检 / 上游有没有更新 / 读下同步文档」→ **先做 §3 第 0 步并发自检**，再执行 skill §2-A
  六步巡检，跑完再汇报。
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
| 面板路由数 | 76 | `grep -rhoE 'HandleFunc\("' internal/panel \| wc -l` |
| 服务端路由数 | 34 | `grep -rhoE 'HandleFunc\("' internal/server \| wc -l` |
| 模型白名单逻辑 | 存在 `NormalizeAllowedModels` | `grep -rn NormalizeAllowedModels internal/livecfg` |
| 账号池选号链路 | `pick.go` / `cooldown.go` / `plan.go` 均在 | `ls internal/pool` |

数字变少 = 收窄 = 不跟，先问清楚再说。

> 基线变更史：54 / 17（2026-10-29 起）→ **63 / 19**（2026-11-02，随上游 41 条提交跟入：
> 代理池与按账号路由 +9 面板路由、Claude Code 计数端点 +2 服务端路由）
> → **65 / 27**（2026-11-06，随上游 40 条提交跟入：账号分组 / 事件日志清空 +2 面板路由，
> 视频端点 + MCP + 水印裁切 +8 服务端路由）
> → **67 / 27**（2026-11-10，随上游 9 条提交跟入：`GET /panel/api/logs/export` +
> `POST /panel/api/accounts/{uid}/scope`）
> → **68 / 27**（2026-11-11，随上游 20 条提交跟入：上游 `65e644e` **撤销了号池分区**，去掉
> `accounts/{uid}/scope`，同时新增账号目录热重载 `accounts/reload` 与 `update/probe`）
> → **70 / 27**（2026-11-13，随上游 8 条提交跟入：运营活动 `GET /panel/api/campaign` +
> `POST /panel/api/campaign/claim`）。
> → **70 / 27 不变**（2026-11-16，A+C 批 47 条跟入，全是行为修复不增路由；模型规则 +5 与
> 混元 3D +3 属 B 批，未跟，跟完才是 75 / 30）。
> → **76 / 30**（2026-11-17，两段共 18 条跟入：第一段 10 条行为修复维持 70 / 27；第二段 8 条
> 补齐 B 批与新功能——模型规则 +5、混元 3D +3，另加 `GET /panel/api/tasks/live` +1 与
> `/status` 积分语义。注：11-16 预估的 75 / 30 少算了任务中心实时卡这一条）。
> → **76 / 34**（2026-11-23，随上游 27 条提交跟入：官方语音端点 `POST /v1/audio/speech` +
> `POST /v1/audio/transcriptions`、DeepSeek 官方工具形状别名 `POST /v1/tools/search` +
> `POST /v1/tools/fetch`，服务端 +4；面板路由不变。另新增配置键 `login_bind_proxy` /
> `tts_enabled` / `asr_enabled`，后者两项默认 false、关着回 403）。
>
> ⚠️ 计数口径是**出现次数**（含测试文件里的 httptest mux），不能单看数字判收窄。
> 判「有没有被砍能力」用路由集合差集：
> `comm -23 <(git grep -hoE 'HandleFunc\("[^"]+"' local -- internal/panel internal/server | sort -u) <(git grep -hoE 'HandleFunc\("[^"]+"' main -- internal/panel internal/server | sort -u)`
> 输出为空 = 本地没有上游缺的路由，再用 `comm -13` 看新增。

## 3. 同步动作

```bash
# 0. 并发自检（多会话 / automation 同时跑同一仓库会互相覆盖文档、重复部署）
git fetch origin
git status --porcelain                      # 必须为空
git log -1 --format='最近提交 %h %ad %an' --date=format:'%m-%d %H:%M:%S'
#   最近提交在 15 分钟内、且不是本会话产生的 → 有人正在动这个仓库：
#   先 `git pull --ff-only` 对齐到远端，再决定要不要继续；
#   若已经部署过一轮，本轮只补差异，别重复构建。
#   （2026-11-23 实证：两个会话同跑同步，代码没撞（收敛到同一组 port 提交），
#     但 SYNC-LOG/GUIDE 被来回覆盖三轮，且各自部署了一次。）

# 1. 巡检（只读）
bash ~/.workbuddy/skills/wyybs-fork-upstream-sync/scripts/inspect.sh .

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

### 3.1 推不上去 / fetch 卡住时（代理失效）

**2026-11-19 实测**：巡检卡在 `fetch` 无输出，根因是**两层代理同时失效**，不是上游问题：

| 层 | 检查 | 症状 |
|---|---|---|
| 沙箱注入 | `echo $http_proxy` → `127.0.0.1:<随机端口>` | Agent 会话里的本地代理，不通外网，`curl` 秒失败 |
| `.gitconfig` URL 级 | `127.0.0.1:7890`（FlClash） | 机场节点失效时超时；`mihomo`/`FlClash` 进程活着但日志报 `dial … error: context deadline exceeded` |

判断网络本身是否健康：`PowerShell` 里 `Invoke-WebRequest https://github.com` 若返回 200，说明线路没问题，只是代理挂了。
**直连其实不慢**——实测 `ls-remote` 2.3s、`curl` 8.9s，比走挂掉的代理强。

```bash
# 本机：清掉两层代理直连（GIT_CONFIG_* 只对本次调用生效，不改 .gitconfig）
env -u http_proxy -u https_proxy -u HTTP_PROXY -u HTTPS_PROXY \
  GIT_CONFIG_COUNT=3 \
  GIT_CONFIG_KEY_0="http.https://github.com.proxy" GIT_CONFIG_VALUE_0="" \
  GIT_CONFIG_KEY_1="http.https://api.github.com.proxy" GIT_CONFIG_VALUE_1="" \
  GIT_CONFIG_KEY_2="http.https://raw.githubusercontent.com.proxy" GIT_CONFIG_VALUE_2="" \
  git fetch upstream --prune --tags
```

**🔴 2026-11-24 更正：22 一直能直连 GitHub，之前「拉不到」全是代理挡路。**

停用 mihomo、清掉 `/root/.curlrc` 的 `proxy` 与 `.gitconfig` 的 URL 级代理后实测：
`git ls-remote` **1.16 秒**、`git pull` 一次成功。所以 22 **不需要**任何代理就能拉代码。

根因：以前所有「直连测试」都被 `.curlrc` 的 `proxy = "http://127.0.0.1:7890"` 接管了
（`--noproxy` 只是"跳过代理的主机列表"，不是禁用代理，得用 `-x ""` 才真直连），
于是测出来的是"代理不通"，被误读成"22 没有直连出网"。**这个误判持续了 5 天，
期间浪费了三次 bundle 兜底和一轮代理修复。**

现状：mihomo 已停用（`docker-compose.yml` 里挂在 `profiles: ["proxy"]`，默认不启动）。
22 拉代码直接 `git pull` 即可；网络慢或想跳过 TLS 抖动时再用 bundle：

```bash
# 本地：只打增量（服务器当前 HEAD 之后的部分，通常几十～几百 KB）
git bundle create /tmp/wb2.bundle <服务器当前HEAD>..local
scp -i ~/.ssh/id_wb_222 /tmp/wb2.bundle root@192.168.2.22:/tmp/
# 服务器
cd /home/appsrv/workbuddy2api-panel
git fetch /tmp/wb2.bundle local && git merge --ff-only FETCH_HEAD && rm -f /tmp/wb2.bundle
```

万一哪天确实需要代理（例如拉境外镜像），临时拉起：

```bash
cd /data/docker && docker compose --profile proxy up -d mihomo   # 用完记得 stop
```

**2026-11-23 补充（历史参考——mihomo 已于 11-24 停用，本节只在临时拉 `--profile proxy` 时才用得上）：
另一种挂法——DNS 死锁，别误判成节点失效。**

日志区别很关键：

| 日志 | 含义 | 处理 |
|---|---|---|
| `dial … error: context deadline exceeded` | 节点地址连不上（机场换入口 / 节点挂了） | 换订阅 |
| `dial … error: dns resolve failed: couldn't find ip` | **mihomo 自己解析不出域名** | 改 DNS 配置 |

判据：`docker exec mihomo getent hosts github.com` 能解析、`nc -z 119.29.29.29 53` 通，
就说明容器网络没问题，是 mihomo 的 DNS 模块死锁了。

根因：从 FlClash 客户端抄来的配置里 `dns.respect-rules: true` + `nameserver` 全是 DoH 域名
（`https://doh.pub/dns-query`）。DNS 查询要遵循代理规则 → 走 hysteria2 节点 → 节点域名又要先解析
→ **循环依赖**。客户端有系统解析器兜底，容器里没有。

修法：

```yaml
dns:
  respect-rules: false          # 关键：DNS 不走代理规则
  nameserver: ["223.5.5.5", "119.29.29.29"]   # 用 IP，别再放 DoH
  fallback: [...]               # DoH 留给 fallback，配合 fallback-filter geoip CN 防污染
```

**换订阅时的 4 处容器适配（每次都照做）**：`allow-lan: false → true`、
`external-controller: "" → "0.0.0.0:9090"`、`secret: "" → "wb2api-proxy"`、
`log-level: "error" → "info"`；再补 `respect-rules: false`（本机客户端有系统解析器兜底，容器里没有）。
重启后**等 20–30 秒**再测——端口监听比容器 `Up` 晚一拍，太早测会得到 `000` 假故障。

订阅真失效时的判据（别拿单个 DNS 下结论）：换 3 个公共 DNS 查同一个节点域名，
`223.5.5.5` / `8.8.8.8` / `1.1.1.1` **全 NXDOMAIN**、但根域 `nslookup -type=ns <根域>` 仍有 NS 响应
→ 机场清空了 A 记录，必须换订阅；只有 mihomo 解析不出来而容器里 `getent hosts` 正常 → 才是 DNS 死锁。

## 4. 校验清单（每次必跑）

- [ ] `git status --porcelain` 为空
- [ ] 动手前做过并发自检（§3 第 0 步）：HEAD 最近提交在 15 分钟内时，已先 `fetch` + 对齐远端
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过（442 个测试文件，慢是正常的）
- [ ] 路由计数回到 76 / 34
- [ ] `go.mod` 的 module 路径仍是我们 fork 的（若上游改名需评估）
- [ ] `md/SYNC-LOG.md` 已追加一条，含回滚 tag
- [ ] `origin/local` 与 `origin/main` 均已推送（`main` 只允许 ff；用 `git ls-remote` 复核，别看 fork 页面）

> ⚠️ **子进程用例与 Agent 沙箱**：`internal/panel` 的前端用例要 fork `node.exe`，`internal/updater`
> 的集成用例要 fork 临时编出来的 `test.test.exe`。沙箱策略收紧时它们会整批报
> `fork/exec …: Access is denied`——**那不是回归**。2026-11-09 实测 52 + 4 条全挂；2026-11-10 复测
> **31 包全绿、0 FAIL**，同一台机器同一批代码，差别只在沙箱策略，所以**先复跑再判断**。
> 真需要判回归时的硬办法：`git worktree add ../_check <上一个 pre-sync-tag>`，在那边跑同一个包，
> 失败条数相同即环境问题。

## 5. 记录去哪看

历史（巡检结论、移植明细、已判定不跟项）全部在 `md/SYNC-LOG.md`。
