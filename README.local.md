# workbuddy2api-panel · 二开说明（local 分支）

> 本文件是**本仓库自己的**说明，记录分支策略、构建方式、部署实例与踩坑。
> 上游原始文档仍是根目录的 [README.md](README.md)，保持原样以便与上游合并。

---

## 1. 这是什么

`wyybs/workbuddy2api-panel` 是 [`misakano7545/workbuddy2api-panel`](https://github.com/misakano7545/workbuddy2api-panel) 的 fork。

它做的事：把多个腾讯 CodeBuddy / WorkBuddy 账号收进一个账号池，对外提供 **OpenAI 兼容 API**，并在每次请求时**自动判断用哪个账号的积分**——不靠人手动切号。

选号逻辑（`internal/` 里的 `pool` 实现）：

```
1. 过滤：禁用 / 冷却中 / 熔断 / 在途占满的账号 → 不参与
2. 按模型「实测成本」分层，只保留当前最优层
3. 快过期加权：expiring_soon（默认 168h）内仍有批次的账号，权重 ×3
4. 同层内加权随机：weight = (本号积分 / 候选集最大积分) × 10 + 闲置补偿
   闲置补偿 = min(闲置小时 × idle_weight_per_hour, idle_weight_max)
5. 防惊群：跳过 100ms 内刚被选中的账号
6. 全池健康号都在冷却 → 从软冷却号里选最早到期者顶班
```

### 血缘与风险（务必知道）

```
Sliverkiss/workbuddy2api        ← 原作者已删库
        ↓
linguo2625469/…-panel           ← 2026-10-21 官宣停止维护（上游开始禁反代）
        ↓
misakano7545/…-panel            ← 本仓库的上游，个人自用分支，更新随缘
        ↓
wyybs/…-panel (本仓库)          ← local 分支二开
```

**整条链没有任何官方支持**，上游随时可能收紧导致通道不可用；使用非官方客户端导致的账号风控由使用者自行承担。

---

## 2. 分支模型

```
upstream (misakano7545)  ← 只读，push 已禁用
        │
        ├── main    只做快进，永远等于 upstream/main
        │
        └── local   所有二开提交都在这，跟踪 origin/local
```

| 分支 | 跟踪 | 推送目标 | 规则 |
|---|---|---|---|
| `main` | `upstream/main` | `origin/main` | 只允许 `merge --ff-only`，**不放自己的提交** |
| `local` | `origin/local` | `origin/local` | 随便改、随便提交 |

远端配置：

```bash
origin    https://github.com/wyybs/workbuddy2api-panel.git     # 自己的 fork
upstream  https://github.com/misakano7545/workbuddy2api-panel.git  # 上游（push URL 已置为 DISABLED）
git config remote.pushDefault origin                            # 裸 git push 走 fork
```

---

## 3. 同步上游

双击根目录的 **`sync-upstream.cmd`**，等价于：

```bash
git fetch upstream
git checkout main
git merge --ff-only upstream/main      # 上游若有强推/分叉，这里会直接失败而不是静默合并
git push origin main
git checkout local
git rebase main                        # 把二开提交重放到最新上游之上
```

失败会自动停住并打印 `git status`，不会强改任何东西。

> **rebase 冲突处理**：`local` 的提交会被重放到新上游之上，改到同一处就会冲突。
> 嫌麻烦可以把最后一步换成 `git merge main`（历史多几个合并节点，但不冲突更省心）。

---

## 4. 构建

### 4.1 本地构建（需 Go 1.26.10）

`go.mod` 要求 **go 1.26.10**。本机当前**未安装 Go**，装好后：

```bash
go build -trimpath -ldflags="-s -w" -o wb2api.exe ./cmd/server
./wb2api.exe -config config.json          # 首次启动自动生成 config.json（含随机 api_key）
```

前端资源已 embed 进二进制，单文件自包含，只需保证 `auths/`、`data/` 可写。

### 4.2 国内网络必须先改 GOPROXY

直连 `proxy.golang.org` **必然超时**（实测 `dial tcp 142.251.33.209:443: i/o timeout`）。
本仓库的 `Dockerfile` 已内置：

```dockerfile
ARG GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOPROXY=${GOPROXY}
```

本地编译同理：

```bash
go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
```

> 这是本仓库对上游 Dockerfile 的**唯一改动**，`git pull` / rebase 冲突时保留这三行。

### 4.3 服务器构建（Docker Compose）

```bash
docker compose up -d --build      # 首次约 3~5 分钟（拉 go 镜像 + go mod download + 编译 4 个二进制）
```

两阶段构建：`golang:1.26.10-alpine` 编译 → `alpine:3.20` 运行，最终镜像约 **124 MB**，容器内以 uid `10001` 非 root 运行。

---

## 5. 当前部署实例

| 项 | 值 |
|---|---|
| 主机 | `192.168.2.22`（Ubuntu 24.04，Docker 29.x + Compose v5） |
| 目录 | `/home/appsrv/workbuddy2api-panel` |
| 容器 | `workbuddy2api`（`restart=unless-stopped`，Docker 开机自启） |
| 端口 | `0.0.0.0:7863` |
| 面板 | `http://192.168.2.22:7863/panel/` |
| API | `http://192.168.2.22:7863/v1` |
| API Key | 见服务器 `config.json` 的 `api_key`（**不写进仓库**） |

数据卷（全部在宿主机，删容器不丢）：

```
/home/appsrv/workbuddy2api-panel/data        -> /app/data         状态 / 积分历史 / 请求日志
/home/appsrv/workbuddy2api-panel/auths       -> /app/auths        账号凭证 token
/home/appsrv/workbuddy2api-panel/config.json -> /app/config.json  配置（面板改会写回）
```

### 定时排程（默认全开）

| 时间 | 任务 |
|---|---|
| 09:00 / 21:00 | 签到（末尾自动跑「连登管家」：兑换解锁档位 + 抽完抽奖次数） |
| 09:00 / 21:00 | 猫猫旅行（领养 / 派出 / 领奖） |
| 10:00 | 活跃上报（点亮连登，解锁 `first_buddy`；国际号另跑网页会话腿） |
| 11:00 | 成长任务队列（扫描待办 → accept → 动作 → 达标领奖） |
| 22:00 | token 保活 |
| 23:00 | 夜猫子（23:00–08:00 窗口补足对话） |
| 每 5 分钟 | 余额后台刷新（余额恢复的账号自动解冻） |
| 每 15 分钟 | 补领清扫（只 accept 待接受、只 claim 已达标未领） |

---

## 6. 关键配置

`config.json` 里值得改的几项：

| 键 | 默认 | 含义 |
|---|---|---|
| `pool.prefer_expiring` | `true` | 快过期积分优先消耗（×3 权重），二开的核心开关 |
| `pool.credit_floor` | `100` | 余额低于此值不再接「实测收费」模型，避免把号打穿到 402 |
| `pool.reserve_credits` | `10` | 保留余额，防止用完触发上游提醒 |
| `pool.idle_weight_per_hour` | `0.5` | 闲置补偿，避免总用同一个号 |
| `session_sticky.enabled` | `true` | 同一会话绑定同一账号，多轮对话不跳号 |
| `breaker_threshold` | `3` | 连续失败达阈值熔断 |
| `schedule.checkin_hours` | `[9,21]` | 签到时点 |
| `watchdog.enabled` | `false` | 上游模型目录变更监控（需配 `alerting.webhook_url`） |

---

## 7. 已知坑

1. **`config.json` 必须先在宿主机存在**，否则 Docker 会把挂载点建成目录，容器起不来：
   ```bash
   cp config.example.json config.json
   ```
2. **容器内必须设 `WB2A_LISTEN=0.0.0.0:7863`**，程序默认只绑 `127.0.0.1`。
3. **挂载目录属主必须是 uid 10001**，否则写 `auths/*.json.tmp` 报 permission denied：
   ```bash
   chown -R 10001:10001 auths data && chown 10001:10001 config.json
   ```
4. **`go mod download` 在国内必挂**，见 §4.2。
5. **面板默认没有登录密码**。谁能访问 7863 谁就能操作账号池——务必只在内网暴露，或用 nginx Basic Auth / 来源 IP 白名单收紧。
6. **升级后要重新 build**：`git pull && docker compose up -d --build`，不能只 restart。

---

## 8. 常用命令

```bash
cd /home/appsrv/workbuddy2api-panel

docker compose ps                          # 状态
docker compose logs -f --tail 100          # 选号 / 签到日志
docker compose restart                     # 改了 config.json 用这个
docker compose up -d --build               # 升级重建
docker exec -it workbuddy2api ./signin.sh   # 手动补签
docker exec -it workbuddy2api ./credit.sh   # 查各号积分余量

curl -s http://127.0.0.1:7863/healthz                                  # 健康检查
curl -s -H "Authorization: Bearer $KEY" http://127.0.0.1:7863/v1/models # 模型列表（池空时返回 []）
```

备份（凭证 + 状态 + 配置，换分支/重装时直接拖过去）：

```bash
tar czf /home/appsrv/backup-wb2api-$(date +%F).tar.gz \
  -C /home/appsrv/workbuddy2api-panel auths data config.json
```

---

## 9. 客户端怎么接

在支持自定义模型 / OpenAI 兼容端点的客户端里，加一条：

```
Base URL : http://192.168.2.22:7863/v1
API Key  : <config.json 里的 api_key>
Model    : 加完账号后从 /v1/models 里选
```

这样客户端**永远只登一个账号**，积分由网关在池子里自动挑——会话不会因为切号而丢。
