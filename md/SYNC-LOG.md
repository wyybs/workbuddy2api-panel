# workbuddy2api-panel 同步日志（append-only，只加不改）

## 已判定不跟（提案前先看这里）

| 上游提交 / 范围 | 内容 | 为什么不跟 |
|---|---|---|
| `96a33f1`（2026-10-30） | 删 4 个「上游已下架」任务 + 整个夜猫子排程：`autoActions` 37→33 条、`scheduler` 任务族 6→5（`blackcat.go` 整个删）、配置三键 `blackcat_hours` / `blackcat_enabled` / `night_account_cap` 消失、`upstream` 删 `InNightWindow`/`BlackcatNeed`/`RunNightChats`/`DesktopDesignCanvasSequence` | ~~不跟~~ → **2026-11-02 改判：已跟入**。原判定依据是我们这台在生产用夜猫子（启动日志「夜猫子已启用：[23] 点」）。2026-11-02 复核实证推翻了这个依据：① `data/task-logs.json` 里 `black_cat` / `Model_chat_GLM5.2` / `Buddy_App_QQ` / `create_canvas` 只在 10-29 出现且**全部 `status=skipped`（该账号没有此任务）**，10-30 之后一次都没有；② 排程 23:00 连续三夜（10-30 / 10-31 / 11-01）**零日志**，说明 `BlackcatNeed` 恒 ≤0，夜猫子每夜只是白打一次上游查询；③ `data/credit-history.json` 里 `black_cat` / `night` / `夜猫子` 出现 **0 次**，从未产过积分。即「已启用」只代表开关开着，实际无任务可做。详见 2026-11-02 条目。 |
| `6724698`（2026-11-03） | `docker-compose.yml` 默认改用 `ghcr.io/misakano7545/workbuddy2api-panel:latest` 预编译镜像，把 `build:` 段整段注释掉，只留「源码构建」注释块 | **不跟**。① 直接在 compose 里**写死上游地址**，命中本仓库的 `HARDCODE_PATTERN=misakano7545` 校验项；② 我们生产是自建镜像 + `docker compose up -d --build`（`APP_VERSION` 由 `.env` 注入），跟了等于把生产切到上游镜像，构建流程与回滚链（`docker tag :pre-<时间>`）同时失效；③ 上游改这条的动机是「自建镜像 appVersion 回落 dev → updater 误报」，该问题我们已用上游 `2bf4587`（`current_known`）解决，不需要换镜像。**若将来想用预编译镜像**：本 fork 自己的 `docker-ghcr.yml` 会发 `ghcr.io/wyybs/workbuddy2api-panel`（`${{ github.repository }}` 自动取 fork 名），应当改指自己的镜像并配 GHCR 认证，而不是指上游。 |
| `fdcbdc7`（2026-11-03） | `Dockerfile` 把 `GOPROXY` 改成 `ARG`，**默认值设为官方源** `https://proxy.golang.org,direct`，并删掉首行 `# syntax=docker/dockerfile:1` | **不跟（已等价实现）**。本地 Dockerfile 早就用 `ARG GOPROXY`，默认值反而是 `https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct`（境内构建不卡十几分钟）。上游默认值方向与我们相反，搬过来会拖慢/挂住 22 上的构建。`# syntax` 那行的删除属可选微优化（少一次 frontend 拉取），不影响构建正确性，暂留。 |

## 时间线（新→旧）

### 2026-11-12 10:00 部署到 192.168.2.22（7 条上线）

- 服务器 `837857f → 40f6875`（`git pull --ff-only`；首次撞 `gnutls_handshake()`，重试即通）。
- 镜像回滚点：**`workbuddy2api-panel-wb2api:pre-20261112-0959`**；配置备份 `config.json.bak-20261112-0959`
  （本批无配置变更，备份只为留档）。
- 重建 1m14s，未触发基础镜像源故障。
- 验证：容器 `running/healthy`；`/healthz` 5 个号、cn + global 可服务；`/panel/` 200；`/v1/models` 200；
  3 分钟日志 `WARN|ERROR` 0 条。**长回答专项验证**（本批动了 narration 长度闸门与续写分流）：
  请求 200 / 3.3s，`finish_reason=stop`，正文 211 字且结尾是完整句（「……需要在显存占用与缓存命中收益
  之间做权衡。」），`rescue` / `narration` / `toolonly` 均无触发记录——新闸门没有误伤正常回答。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261112-0959 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或服务器 `git reset --hard 837857f` 后重建。
- **踩坑记录**：`if timeout 120 git pull --ff-only 2>&1 | tail -2; then break; fi` 这种写法里，
  `if` 判的是管道末端 `tail` 的退出码（恒 0），重试循环会在第一次就 `break`，看起来像"成功"其实没拉。
  正确写法是 `git pull > /tmp/pull.log 2>&1` 后判退出码，再 `tail` 看日志。

### 2026-11-12 09:58 跟入上游 7 个提交（签到回执四态 / 会话粘性 / 身份指令清洗），基线不变

- 巡检：上次同步点 `baf9155`（11-11 11:03 部署完），`inspect.sh` 自动 ff `main` → **`8125831`**
  （tag `v2026.11.11.2309`），`local` @ `837857f`，工作区干净。新提交 **7 条（含 1 merge）**，33 文件 +1196/−1031。
- 收窄判定：**零收窄**。路由集合差集为空、无删除文件、无新增配置键，基线维持 **68 / 27 / 12**。
- 回滚 tag：**`pre-sync-20261112-095610`**（= `837857f`）。
- 搬入 6 条 `port:` 提交，**零冲突**：
  - `8125831` 重复签到不再报「签到已完成」，按回执分四态出文案（新增 `upstream/checkin_receipt.go` 177 行）
  - `2b210e9` 请求级错误不再解绑会话粘性——**正是 11-11 那条 `cache miss: prompt=106678` 的根因**
  - `8f9fa3c` 清洗上游注入的身份指令（整句 `when asked about your … identity, identify yourself as …`）
    + 去掉全球号「等签到恢复」的错误承诺
  - `8818458`（merge `cf16478` 的内容由它覆盖）`buddyapp_enter_click` 的 `isFirstPage` 改布尔，
    修掉 Buddy_App 任务永远 0/1
  - `79ba3d6` 等用户拍板的收尾不再救援；narration 加长度闸门；面板分列救援入口
  - `92dadae` 工具调用标记按完整性分流——截断的补完、完整的别续写
- 清洗口径核对：正则要求**完整句式**，源码注释明确写着「单独匹配 `identify yourself as` 会误伤
  `Do not identify yourself as an AI.` 这类合法要求；`<model_identity>` 只删标签、保留内部文本；
  入口有 `strings.Contains` 预检走零分配快路径。**不会误伤正常正文。**
- 部署前采集的线上实况（近 24h）：19 条 `WARN|ERROR`，全是 `travel` / `streak-bonus` 的上游 500
  （各 3–4 条）、2 条 `user-resource` 429 soft_rate、1 条 cache miss——**无一条本地故障**。
  `streak-bonus` 四个号全 500，正对应本批的签到回执改造。
- 校验：`go build ./...` / `go vet ./...` 通过；`go test ./...` **31 包全绿、0 FAIL**
  （panel 48.0s / scheduler 29.5s / server 28.8s / updater 12.4s）；路由 68 / 27 / 12、差集空；
  `git status --porcelain` 空。

### 2026-11-11 11:03 部署到 192.168.2.22（20 条上线，含号池分区回滚）

- 服务器 `e58bc29 → 8474c01`（`git pull --ff-only`；首次撞 `gnutls_handshake()`，重试一次即通，老毛病）。
- 镜像回滚点：**`workbuddy2api-panel-wb2api:pre-20261111-1101`**；配置备份 `config.json.bak-20261111-1101`。
- `config.json` 删掉 `pool.separate_text_media`（上游已撤销该功能）。主配置解析不走 `DisallowUnknownFields`，
  留着也不会启动失败，但删掉更干净，避免下次误以为还有这个开关。
- 重建 1m29s（这次基础镜像源正常，未再触发 503/429）。
- 验证：容器 `running/healthy`；`/healthz` 5 个号、cn + global 可服务；`/panel/` 200；`/v1/models` 200
  且响应里出现 **`promo_label`**（`b64773e` 生效）；新路由 `POST /panel/api/accounts/reload`、
  `POST /panel/api/update/probe` 未鉴权均 **401**（存在且走鉴权）；**已撤销的 `POST /panel/api/accounts/{uid}/scope`
  带 api_key 访问为 404**，确证路由已随回滚删除；配置回读 `separate_text_media=None`、
  `selection_strategy=weighted`、`window_probe_enabled=True`；3 分钟日志 `WARN|ERROR` 0 条；内存 15.6MiB；
  真实对话 200 / 1.73s / 返回「收到」。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261111-1101 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或服务器 `git reset --hard e58bc29` 后重建。

### 2026-11-11 11:00 跟入上游 20 个提交（回滚号池分区 / 签到补签 / 依赖漏洞 / 账号热重载），零收窄

- 巡检：上次同步点 `592fb01`（11-10 10:36 部署完），本次 `inspect.sh` 自动 ff `main` → **`baf9155`**
  （tag `v2026.11.11.832`），`local` @ `e58bc29`，工作区干净。新提交 **20 条（含 1 merge）**，72 文件 +4989/−1637。
- **上游撤销了 11-10 刚跟入的号池分区**（`65e644e`）：理由与我们 11-10 的建议一致——分池要运维逐号标
  text/media，代价（人工维护 + 号少时两边更容易 503）大于收益。撤销 21 文件、净删 607 行：
  `pool/scope.go` 与测试、`panel/account_scope_test.go`、`POST /panel/api/accounts/{uid}/scope`、
  `pool.separate_text_media` 全部删除，`pick` 恢复三参数签名。**额度窗口探针与代理订阅导入保留**。
  回看 11-10：那次跟入时 `separate_text_media` 设的是 `false`、从未启用，所以本次回滚零损失。
- 收窄判定：**零收窄**。路由集合差集为空（脚本报的「本地独有 scope 路由」是上游删掉了它）。
  基线 **67 / 27 → 68 / 27**（撤 1 条、增 2 条：`POST /panel/api/accounts/reload`、`POST /panel/api/update/probe`）。
- 回滚 tag：**`pre-sync-20261111-105733`**（= `e58bc29`）。
- **`go.mod` 同步以来首次变动**：`klauspost/compress v1.17.4 → v1.20.2`，修可达漏洞 GO-2026-6865（`849d338`）。
  module 路径未动。
- 搬入 **19 条 `port:` 提交，零冲突**：
  - 号池与计费 4 条：`1adfa52`（签到遇「请稍后重试」10 / 40 分钟后自动补签）、`cfef70d`（国际版到期批次按
    PackageCode + 周期聚合）、`cf2b80f` / `d011fc8`（计费用例锚到日界，修掉 CST 00:00–12:00 必红）
  - 续写与救援 4 条：`baf9155`（「已写完的答复」不救）、`2e30d7d`（扩大「说一句就停」救援覆盖 + 面板统计）、
    `910d851`（安全续写正文上限 240B → 2000B）、`e8df35b`（正文里的 DSML 文本不触发续写）
  - 上游兼容 2 条：`afb5f22`（一个 data 行多 JSON 时拆分）、`8df99ac`（等首帧期间补 SSE 注释帧）
  - 账号与热重载 3 条：`9e5c2cd`（账号目录热重载 + 更新代理/镜像/失败归因）、`c54c517`（落盘与热重载串行化）、
    `4f6adc0`（`reload_delete_confirm>1` 时永不删号；merge `a439202` 的内容由它覆盖）
  - 面板与代理 2 条：`cfac370`（单选框宽度）、`7366740`（订阅 proxies 流式数组不再被丢）
  - 其他 4 条：`d7ffe42`（缺函数名告警不算已发出的调用）、`b64773e`（`/v1/models` 透出 promo_factor /
    promo_credits / promo_label）、`65e644e`（回滚分区）、`849d338`（依赖升级）
- 校验：
  - `go build ./...` / `go vet ./...` 通过（首次构建会下载 compress v1.20.2）。
  - `go test ./...` **31 包全绿、0 FAIL**（panel 42.7s / scheduler 28.8s / server 26.4s / updater 11.9s）。
  - 路由 68 / 27、白名单 12、`git status --porcelain` 空。
  - 分区残留核对：`pool/scope.go` 已不存在，代码 / 配置 / 前端 grep `separate_text_media` 为 **0**；
    仅 `md/SYNC-LOG.md` 与本机 `.workbuddy/memory` 的历史记录里还提到（append-only，正常）。
  - 生产 `config.json` 里遗留的 `pool.separate_text_media` 会被忽略（主配置解析不走
    `DisallowUnknownFields`，只有 `panel/proxies.go` 的请求体用），部署时顺手删干净。

### 2026-11-10 10:36 部署到 192.168.2.22（9 条上线；基础镜像源两次故障）

- 服务器 `d910a60 → 33b0007`（`git pull --ff-only`，10 提交，无冲突）。
- 镜像回滚点：**`workbuddy2api-panel-wb2api:pre-20261110-1027`**；配置备份 `config.json.bak-20261110-1027`。
- `config.json` 补 8 个新键（值即采纳的默认值；`schedule.makeup_cards_per_round` 原本已是 0）：
  `logging.request_archive_context=false` / `panel_log_enabled=true` / `panel_log_retention_days=7` /
  `panel_log_max_mb=50`、`schedule.window_probe_enabled=true`、
  `pool.selection_strategy="weighted"` / `separate_text_media=false`。diff 只新增这三段，JSON 复验合法。
- **第一次 build 失败**：`docker compose up -d --build` 在 `FROM alpine:3.20` 处报
  `docker.1ms.run … 503 Service Unavailable`；直接 `docker pull` 又撞 `docker-0.unsee.tech … 429`。
  处理方式：**不动 daemon.json**（改它要重启 docker，会带走所有容器），改为从 `docker.m.daocloud.io`
  拉 `alpine:3.20` 与 `golang:1.26.10-alpine`，`docker tag` 回标准名，第二次 build 3m52s 成功。
  期间旧容器全程 `Up (healthy)`，服务零中断。
- 验证：容器 `running/healthy`；`/healthz` → `{"healthy":5,"realm_servable":{"cn":true,"global":true}}`；
  `/panel/` 200；`/v1/models` 200；`/panel/api/logs/export` 未鉴权 **401**（新路由存在且走鉴权）；
  6 项新配置经 `/panel/api/config` 回读全部生效；3 分钟日志 `WARN|ERROR` 0 条；内存 13.9MiB；
  真实对话 200 / 1.82s / 返回「收到」，`[budget] max_tokens 64 -> 8192` 抬额机制正常。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261110-1027 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或服务器 `git reset --hard d910a60` 后重建。
- 运维经验：**国内镜像源不稳时的标准解法**——`docker pull docker.m.daocloud.io/library/<name>:<tag>`
  再 `docker tag` 回标准名，比改 daemon.json 安全（不重启 docker、不影响别的容器）。2026-11-10 实测可用。

### 2026-11-10 10:30 跟入上游 9 个提交（额度窗口探针 / 号池分区 / 选号策略 / 运行日志落盘），零收窄

- 巡检：上次同步点 `24d0b61`（11-09 10:40 部署完），本次 `inspect.sh` 自动 ff `main` → **`592fb01`**
  （tag `v2026.11.9.1131`），`local` @ `d910a60`，工作区干净。新提交 **9 条、零 merge**，68 文件 +5969/−949。
- 收窄判定：**零收窄**。本地独有路由为空；上游新增两条面板路由 `GET /panel/api/logs/export`、
  `POST /panel/api/accounts/{uid}/scope`。基线随之上移 **65 / 27 → 67 / 27**（服务端 27、白名单 12 不变）。
- 回滚 tag：**`pre-sync-20261110-102404`**（= `d910a60`）。
- 搬入 9 条 `port:` 提交，**零冲突**：
  - 号池调度 3 条：`592fb01`（额度窗口探针 + 号池分区 + 代理订阅导入；`internal/pool` 新增
    `scope.go` / `strategy.go` / `window.go` / `egress.go`）、`18a6776`（选号策略可配 + 运行日志落盘导出 +
    上游劣化滑窗告警）、`bfdba51`（直连传输层失败记到出口，不再按轮转逐号记连败）
  - 续写与工具轮 3 条：`c2003af`（正文停在预告句就收尾时续写）、`1037397`（工具轮「说一句就停」试续一次、
    只采纳工具调用）、`44bc059`（工具调用缺函数名被丢弃时留 WARN）
  - 面板与上游 2 条：`c5ce29b`（账号导入支持多文件、选中先预览再写入）、`d78131b`（资料接口空昵称不再静默成功）
  - 诊断 1 条：`4bdbd60`（narration/toolonly 记录尾部片段与续写结果）
- 新增配置键（默认值即采纳值）：`logging.panel_log_enabled=true` / `panel_log_retention_days=7` /
  `panel_log_max_mb=50` / `request_archive_context=false`、`schedule.window_probe_enabled=true` /
  `makeup_cards_per_round=0`、`pool.selection_strategy="weighted"` / `separate_text_media=false`。
  **分区默认关、策略默认 weighted = 分区前的行为逐字不变**；探针只在「429 已登记重置时刻」到点发一次
  `Hi` / 16 token，刻意不做按账号×模型扇出的配额扫描。
- 校验：
  - `go build ./...` / `go vet ./...` 通过。
  - `go test ./...` **31 包全绿、0 FAIL**（panel 43.8s / server 26.4s / scheduler 23.5s / updater 12.0s）。
    11-09 那批在沙箱里挂掉的 52 + 4 条子进程用例这次全部通过——同一台机器、同一批代码，差别只在沙箱策略，
    已在 GUIDE §4 把这条说明改写为「先复跑再判断」。
  - 路由 67 / 27、白名单 12、`git status --porcelain` 空、`go.mod` module 路径未变。
  - `git grep misakano7545` 只剩 `go.mod` 的 module 路径。

### 2026-11-09 10:40 部署到 192.168.2.22（57 条上线）

- 服务器 `3bfdd2c → 68303fe`（`git pull --ff-only`，50 提交，无冲突）。首次 `git fetch` 撞
  `gnutls_handshake() failed`，重试即通（mihomo Up 6 天、`curl github` 200，属 GnuTLS 偶发抖动）。
- 镜像回滚点：**`workbuddy2api-panel-wb2api:pre-20261109-1038`**（= 旧 `cb87e90a60b0`）。
- `config.json` 三处调整（备份 `config.json.bak-20261109-1038`）：`restriction_hold_seconds` **300**（默认 1800
  会把 3 个号的池子在 2 个号中 11140 后整停 30 分钟）、`server.max_payload_bytes` 显式写 **67108864**、
  `thinking.default_effort` **max → high**。改完 diff 只动这三行，JSON 复验合法。
- 重建：`docker compose up -d --build` 1m27s → 容器 Recreated / Started → `running/healthy`。
- 验证：`/healthz` → `{"healthy":5,"realm_servable":{"cn":true,"global":true},"total":5}`；`/panel/` 200；
  `/v1/models` 200（0.24s）；真实对话 200（1.45s，返回「收到」）；近 5 分钟 `WARN|ERROR` **0 条**；
  内存 14MiB；面板 `/panel/api/config` 回读确认三处配置已生效。
- 新能力实证：日志打出 `[budget] max_tokens 64 -> 8192（effort=high 的输出预算下限……）`，
  请求记录行带 `effort=high`、`cache=99%`。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261109-1038 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或服务器 `git reset --hard 3bfdd2c` 后重建。

### 2026-11-09 10:25 跟入上游 57 个提交（长文续接 / 输出预算 / 号池自救 / 面板 reqlog 区分），零收窄

- 巡检：上次同步点 `dab1b08`（11-06 10:53），本次 `inspect.sh` 自动 ff `main` → **`24d0b61`**
  （最新 tag `v2026.11.9.827`），`local` @ `9386585`，工作区干净。新提交 **57 条**（48 单提交 + 9 merge），
  110 文件，+8968 / −328。
- 收窄判定：**零收窄**。路由集合双向差集为空（`comm -3` 无输出）、无删除文件、无新增写死
  `misakano7545` 的地址。面板路由 65 / 服务端路由 27 / `NormalizeAllowedModels` 12，与上游同值。
- 回滚 tag：**`pre-sync-20261109-101507`**（= `9386585`，`git reset --hard` 可整段退回）。
- 搬入 **48 条 `port:` 提交，零冲突**：
  - 长文续接与输出预算 8 条：`2b9b773`（输出预算下限）、`e54c6e0`（救援段继承下限）、`34a3ae6`（length 处续接）、
    `29db878`（闸门从「最多 3 段」改成累计上限）、`562d0d8`（思考吃光也续写）、`6e7fe62`（客户端 ≥32k 直接续接）、
    `84a48d9`（断流补截断终态）、`07f8ae0`（思考完直接收尾时救援，空回复不再 502）
  - 号池健康 3 条：`4830686`（11134 供应商过载不罚号、保会话粘性）、`f987bac`（11151 归请求级；
    Kimi/GLM 补 reasoning 镜像）、`4fb7cb4`（盲盒消费 + 10105 活跃会话自救）
  - 凭据与导入 5 条：`c7cc216`（新版桌面端 `$wbEncrypted` 自动解密）、`97d3f25`（回写只封读入时确实加密的字段）、
    `e36ad07`（覆盖凭据前先求证新凭据还活着）、`585ae7f` / `cd88450`（cockpit 导入兼容嵌套/驼峰/字典形态）
  - 面板与 reqlog 7 条：`6671b8d` / `efe64fd` / `b674f71`（`model_eff` 全链路区分降级前后模型）、
    `4f37ffa` / `69248a6` / `1b3361a` / `f6335ce`（控件尺寸、配置页错位、筛选偏好、余额筛选器绑定）、
    `e4e6f5c`（兜底 system 只写行为约定、不声明身份）
  - 加固 3 条：`b494426`（入站体上限 64MiB + 11140 风控连坐保护）、`c6f953a`（慢上传可见性 + 在途名额回归）、
    `81b806c`（成品直链只认产物域、不跟随重定向）
  - 性能 3 条：`4b0a3b4`（请求体只解析一次）、`fe51fd6`（去掉 RE2 折叠，省 ~20ms CPU/请求）、
    `89c9b7d`（流式按缓冲批量 flush）
  - 诊断与脚本 4 条：`b42712c`（`WB2API_PPROF`，默认关）、`478d2aa`（抓「声明工具却只说一句」样本，默认关）、
    `24d0b61`（`cache_audit.py`）、`7028676`（构建产物不入库）
  - 其余 5 条：`ffca1da` / `1df3b83` / `877324f`（默认档统一到 high）、`d165e6e`（只发 `user_id` 也派生会话键）、
    `771c887`（`<think:SUF>` 拆回 thinking 块）、`29d9879`（`/v1/search` 卡片改等价检索链接）、`7297824`（WARN body 200→512）、
    `f0ca4d3` / `fa5d614`（image_generation 积分闸与断连语义）、AutoBot 4 条、`df988aa`（README Windows 常驻）
- **9 个 merge 节点未单独搬**：逐条比对 `git diff <merge>^1 <merge>` 的文件清单后确认，48 条单提交已完整覆盖其内容
  ——搬完 `git diff main..HEAD` 的代码层**只剩 `Dockerfile` + `docker-compose.yml`**（= 11-06 刻意剥离的两条）。
  11-06 那次「merge 独有内容缺失」本次**没有复现**，但「pick 完跑全量测试」这条做法继续保留。
- 剥离：**无新增**。`6724698`（compose 默认 GHCR 上游镜像）、`fdcbdc7`（Dockerfile GOPROXY 官方源）继续不跟。
- 校验：
  - `go build ./...` / `go vet ./...` 通过。
  - `go test ./...`：**28 包 ok**；`internal/panel` 52 条、`internal/updater` 4 条失败，错误全是
    `fork/exec …: Access is denied`——沙箱禁止 fork `node.exe` 与临时编出的 `test.test.exe`，不是代码问题。
    **已用 worktree 对照上一 tag 证伪回归**：同步前该包 50 条失败 → 同步后 52 条，增量正是本次新增的 2 条 node 用例。
  - 绕过沙箱直验前端：`node --check internal/panel/app.js` 通过。
  - 路由 65 / 27、白名单 12、`git status --porcelain` 空、`go.mod` module 路径未变。
  - `git grep misakano7545` 仅命中 module 路径、import、`appRepository` / updater 默认仓、测试夹具（历史既有，非本次引入）。
- 文档修正：`SYNC-GUIDE.md` §4 清单里过期的 `63 / 19` → `65 / 27`、`222 个测试文件` → `282`，
  并补了「沙箱里跑不了的用例怎么判回归」的说明。
- **部署待办（本次未做）**：① 服务器 `data/config.json` 把 `restriction_hold_seconds` 从默认 1800 调到 **300**
  ——我们只有 3 个号，15 分钟内 2 个号中 11140 就全池停 30 分钟；② `thinking.default_effort` 显式写 **high**
  （上游实测 max 正文 1906 token vs high 3236）；③ 按老规矩先 `docker tag …:pre-<时间>` 再 `up -d --build`。
  → **已于 10:40 部署完成，见上一条**。

### 2026-11-06 10:53 跟入上游 38 个提交（视频 + MCP + 水印裁切），退役本地 updater 补丁

- 巡检：上次同步点 `abca6fc`（11-02 11:20），本次 `inspect.sh` 自动 ff `main` → **`dab1b08`**
  （上游最新 tag `v2026.11.5.2306`），`local` @ `5bca715`，工作区干净。新提交 40 条（无 merge），
  83 个代码文件，上游净增约 8400 行。
- 收窄判定：**零收窄**。路由集合差集实测——上游新增 10 条、**本地独有 0 条**：
  `POST/GET /mcp`、`POST /v1/videos`、`/v1/videos/generations`、`/v1/videos/tasks`、
  `GET /v1/videos/{task_id}`、`GET /v1/videos/{task_id}/content`、`POST /v1/images/strip-watermark`、
  `POST /panel/api/accounts/{uid}/group`、`POST /panel/api/request_logs/clear`。
  （`inspect.sh` 报的 65/27 vs 63/19 是含测试文件重复注册的出现次数，不是净增；已在 GUIDE §2 补口径说明。）
- 回滚 tag：**`pre-sync-20261106-105300`**（= `5bca715`，`git reset --hard pre-sync-20261106-105300` 可整段退回）。
- 搬入 **38 条 port 提交 + 1 条收尾**（`git cherry-pick -x` 逐条，重写为 `port:` 前缀，保留溯源行）：
  - 修复 25 条：`24b8f39`（默认档位误落 off 关思考）、`02b941e`（流式连接无法复用）、
    `5866d6f`（#149 成长任务 24 次列表查询撞 429）、`a8be846` / `66e5344`（客户端断开归类）、
    `b08e606`（心跳帧吞 Responses 事件）、`2af7de6`、`8d48ee5`（#150 代理全停回落直连）、
    `8f761c1` / `e65f786`（reqlog）、`6d534c5`（#155 幽灵账号）、`77d4762`、`526e959`、
    `6ddbaf8` / `f6c4a45`（企业版 / 签到判据）、`f595ed1` / `5169381`（水印裁切）、
    AutoBot 6 条（`9a122b6` / `edcfdf8` / `ac22925` / `d8cb1db` 等）
  - 新能力 9 条：`5fba8aa`（出视频端点 + MCP 端点，5 个工具）、`8a86156`（可选水印裁切，**默认关闭**）、
    `6df19e5`（OpenAI 资源形状 / `b64_json`）、`dab1b08`（托管 `image_generation` 工具）、
    `f75b91c`（国际版出片）、`e6dd088`（账号分组 + 事件日志清空）、`7be4c3b`（#147 保留积分按域）、
    `dd01a52`（tools 描述脱敏 + 轮询指纹跳过）、视频台账 4 条（`b119835` / `51d8173` / `fbb01f7` / `aa9d58d`）
  - 测试 2 条：`e5afc3f` / `d596fa7`；chore 2 条：`40d98f9`（删 404 上游仓）、`e1436d1`（AutoBot 停每日定时）
- **剥离 2 条**（判定见「已判定不跟」表）：`6724698`（compose 默认 GHCR 上游镜像）、
  `fdcbdc7`（Dockerfile GOPROXY 默认官方源）。
- **本地补丁退役**：`110d0e6`（`local_build` 字段修「本地构建误报有新版本」）被上游 `2bf4587`
  （`current_known` 字段，同逻辑但测试覆盖更全）取代。冲突时四个文件（`internal/updater/updater.go`、
  `updater_test.go`、`internal/panel/update_test.go`、`app.js`）整体取上游版；
  `git grep local_build/LocalBuild` 已零残留。**注意前端文案随之变成上游措辞**，
  原先那句「本地构建（未注入版本号）· 跳过版本检查」不再存在。
- 冲突处理共 2 处：
  1. `2bf4587`：4 个文件如上，取上游（theirs）。
  2. **merge 节点整合内容缺失**：`24b8f39` 走 PR 分支合并，其 `app.js` 侧 41 行
     （`CFG_MAP` 的 `thinking_default_effort` 条目、`loadConfig` 回填后与下拉 change 时调
     `refreshEffortDowngradeNote`、以及该函数本体）**只存在于 merge 节点**，`2bf4587` 单提交的树里没有。
     直接 pick 单提交后 `TestConfigFormMatchesCFGMap` 报「`thinking_default_effort` 在 CFG_MAP 里没有条目
     （保存时会被静默丢弃）」。收尾提交 `d799009` 按 `main` 的 `app.js` 补齐。
     **教训：上游有 merge 提交时，被合并分支的单提交树可能不含主线文件的最新状态，pick 后必须跑全量测试。**
- 校验：`go build ./...` / `go vet ./...` 通过；`go test ./...` **27 包全绿、0 FAIL**（含新包
  `internal/media`）；面板路由 **65** / 服务端路由 **27**（= 上游同值）；`NormalizeAllowedModels` 12；
  `git status --porcelain` 空；`go.mod` module 路径未变。
- **代码层残差**：`git diff main..local -- '*.go' '*.js' '*.html' 'go.mod' 'go.sum' '*.yml' '*.json'`
  只剩 `Dockerfile` + `docker-compose.yml` 两个文件——即上面刻意剥离的两条，
  **除这两条外与上游完全对齐**。其余全量残差均为本地独有文件
  （`.fork-sync.*`、`README.local.md`、`README.md` 顶部标注、`md/SYNC-*`、`sync-upstream.cmd.disabled`）。
- `misakano7545` 检查：命中仅 module 路径、import、测试夹具，以及 `cmd/server/main.go`
  与 `internal/updater/updater.go` 的默认 `appRepository`（上游默认值，历史既有，未本次引入）。
- 新端点的鉴权已核对：`POST /mcp`、三条 videos 路由全部走 `withAuth`，`GET /mcp` 直接 405，无未鉴权入口。
- 推送：`origin/local` `5bca715..3bfdd2c`（含 39 个 port 提交 + 1 个 local 日志提交）。

### 2026-11-06 11:07 部署到 192.168.2.22（38 提交 + 视频/MCP 上线）

- 服务器 `/home/appsrv/workbuddy2api-panel`：`5bca715 → 3bfdd2c`（`git fetch origin local` +
  `git merge --ff-only`，未走 bundle；mihomo 代理在位，GitHub 拉取一次成功）。
- 镜像回滚点：**`workbuddy2api-panel-wb2api:pre-20261106-1108`**（上一版 `:pre-20261102-1115` 仍在）。
- 重建：`docker compose up -d --build` → 容器 Recreated / Started，**`Up (healthy)`，`RestartCount=0`**。
- 部署前基线（对照用）：`/healthz` `{"healthy":4,...}`、`/panel/` 200、RestartCount 0、
  日志有 `WARN: [continue] 段无 finish 结束（疑似上游断流，按截断终态降级）` 一条。
- 部署后验证：
  - `/healthz` → `{"healthy":4,"realm_servable":{"cn":true,"global":false},"total":4}`；`/panel/` **200**；
    `WARN|ERROR` **0 条**；启动恢复 132 个用量桶 / 656 条积分流水 / 4 个账号全部在线。
  - 启动日志 **`[upstream] TLS 指纹复刻已启用（JA3 71dc8c533dd919ae9f4963224a4ba8fd / JA4 t13d0a1200_5d04281c6031_b76c5ee2507d）`**
    在；`夜猫子` 命中 **0**（删净）。
  - 真实流量：#001–#005 全 `200 / out=success`，**非 200 零条**（`src=192.168.2.127` 正常客户端）。
  - **新端点实测**：`POST /mcp` 无 key → **401**（鉴权包住）；带 key `initialize` 正常握手并返回
    `tools` capability + 中文 instructions；`tools/list` → `list_generation_models` / `generate_image` /
    `edit_image` / `generate_video` / `get_video_task` 五个工具；`GET /mcp` → **405**；
    `POST /v1/images/strip-watermark` → **403 `watermark_strip_disabled`**（默认关，符合设计）；
    `/v1/models` 里 `video_generation:true` **2 个**（视频模型已进目录）。
  - 版本：`/panel/api/update` → `{"current_version":"dev","current_known":false,
    "latest_version":"2026.11.5.2306",...}` —— **`current_known:false` 说明 `2bf4587` 生效**，
    dev 构建不再声称「发现新版本」，只报查到的最新 release。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261106-1108 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或代码侧 `git reset --hard pre-sync-20261106-105300` 后重建。
- 待观察：① 视频端点与 `generate_video` 会真实扣积分，注意 `data/credit-history.json` 消耗速率；
  ② `WARN: [continue] 段无 finish 结束` 这类日志在本轮 `b08e606` / `a8be846` / `66e5344` 之后
  是否消失（部署后 0 条，需更长时间观察）；③ 面板仍显示 `dev`，如需真实版本号在 22 的 `.env`
  写 `APP_VERSION=2026.11.6.1053` 再重建。

### 2026-11-02 11:20 补跟上 4 条（代理池 TLS 隧道修复）+ 二次部署

- 触发：上一轮同步完成后 6 分钟，上游又推 6 条（`8903487 → abca6fc`）。其中 `673a446` 直接修刚上线的
  代理池：SOCKS 分支下 https 请求**整条不经代理、从本机 IP 出网**，HTTP/HTTPS 代理下**指纹与头顺序全部失效**
  （代理池 Transport 克隆自 `upstream.New()` 时连直连的 `DialTLSContext` 一起继承）。属必须跟。
- 回滚 tag：**`pre-sync-20261102-111337`**（= `09fc5f8`）。
- 搬入 4 条（`f909de7` / `410398f` 两条 PR merge 内容分别等同 `673a446` / `76727b8`，不重复搬）：
  - `673a446` port: fix(proxypool): 经代理的 HTTPS 先建隧道再走指纹握手（#135）
  - `76727b8` port: fix(panel): 按域降级键命中与锁池表「已路由至」标注
  - `bbf2bb9` port: feat(autobot): 给 CI 的 Hermes 装上 humanize 技能
  - `abca6fc` port: fix(autobot): 分析阶段也带 GH_TOKEN
- 冲突 1 处：`bbf2bb9` 的 `.gitignore` —— 上游加 `!.github/autobot-skills/**/*.md`，我们本地有
  `!README.local.md` + `!md/SYNC-*.md`，**两边都留**。
- 校验：build / vet 通过；`go test ./...` 27 包全绿；`internal/proxypool` 含新增
  `tlsover_test.go`（173 行 TLS-over-proxy 用例）通过；代码层与上游差异仍是那 4 个文件的更新误报修复。
- 部署：`origin/local` `09fc5f8..575fc40` 推送；服务器 `git fetch` 首次无输出（GitHub 抖动），
  重跑即通（未用 bundle）；镜像回滚点 **`pre-20261102-1115`**；
  `docker compose up -d --build` → `Up (healthy)`，`RestartCount=0`，
  `/healthz` `{"healthy":2,...}`、`/panel/` 200、`WARN|ERROR` 0 条；
  curl 真实对话 `cn:fast-model` 返回 `content:"OK"`，近 5 分钟 3 条请求全 `out=success`、非 200 零条。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261102-1115 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或 `git reset --hard pre-sync-20261102-111337` 后重建。

### 2026-11-02 10:50 跟入上游 43 个提交，并改判 `96a33f1`（删夜猫子）跟入

- 巡检：上游新增 3 个 tag（`v2026.10.30.1049` / `v2026.10.31.845` / `v2026.11.1.905`），
  `inspect.sh` 自动 ff `main` → **`8903487`**（快进前落后 43 个提交）；`local` @ `4b5f304`，工作区干净。
- 回滚 tag：**`pre-sync-20261102-104803`**（= `4b5f304`，`git reset --hard pre-sync-20261102-104803` 可整段退回）。
- 搬入 **45 个 port 提交**（`cherry-pick -n` + 重写为 `port:` 前缀，保留 `(cherry picked from commit …)` 溯源）：
  - 上游 43 条里跟 43 条、跳过 0 条（原先圈定跳过的 `cbc8aaa` / `5caddab` 两条「文案收尾」在删掉夜猫子后
    反而成了必须补的收尾，一并跟入，见下）；
  - `96a33f1` 改判跟入（依据见本节末尾「改判实证」）；
  - 1 条本地收尾（`4c6ba8c`）：`autotask.go` 跳过逻辑的注释还挂着已删任务名，直接取上游版。
- 冲突处理（共 4 处，全部手解）：
  - `2d23068`：`cmd/server/main.go` 与 `internal/scheduler/scheduler.go` —— 上游删 `BlackcatDisabled` 又加
    `WebActivityDisabled`，此步**两个都留**（夜猫子当时还在）。
  - `96a33f1`：同两处反向 —— **取上游**：删 `BlackcatDisabled`，保留 `WebActivityDisabled`。
  - `b2df95d` 零冲突：发版条件由「`v*` 标签 或 workflow_dispatch」收紧为**只认 `v*` 标签**，
    与 AGENTS.md「工作流看到 v* 才创建 Release」一致，直接接受。
- 校验：`go build ./...` / `go vet ./...` 通过；`go test ./...` **27 包全绿、0 FAIL**；
  面板路由 **63** / 服务端路由 **19**（= 上游同值，随 `56bef12` + `2d23068` + `8903487` 增长，非收窄）；
  `NormalizeAllowedModels` 12；`autoActions` 条目 37→33（随 `96a33f1`）。
- 代码层残差：`git diff main..local -- '*.go' '*.js' '*.html' 'go.mod' 'go.sum'` 只剩 **4 个文件**
  （`internal/updater/{updater.go,updater_test.go}`、`internal/panel/{app.js,update_test.go}`），
  即我们自己的 `110d0e6` 修「发现新版本」误报那一条（`local_build`）。**除这一条外与上游完全对齐**，
  以后再同步的冲突面基本归零。
- 改判实证（推翻 10-30 的不跟理由）：
  - 生产 `config.json`：`blackcat_enabled=true`、`blackcat_hours=[23]`、`night_account_cap=0`，开关确实开着；
  - 但 `data/task-logs.json` 里 `black_cat` / `Model_chat_GLM5.2` / `Buddy_App_QQ` / `create_canvas`
    只在 **10-29** 出现，且 3 次全是 `status=skipped`（「该账号没有此任务」），10-30 之后零出现；
  - 排程 23:00 连续三夜（10-30 / 10-31 / 11-01）**零日志** → `BlackcatNeed` 恒 ≤0，夜猫子每夜只白打查询；
  - `data/credit-history.json` 中 `black_cat` / `night` / `夜猫子` **0 次**——从未产出积分；
  - 结论：「已启用」只说明开关开着，任务已从上游下架，代码是死重量；且夜猫子是「整任务都是真实模型对话」，
    深夜全池同 IP 齐做真实对话本身是最好认的机器人形状，删掉反而降低指纹风险。
- 推送与部署：`origin/local` `4b5f304..a6ae90b` 已推；生产已上线（见下一条 11:05 记录）。

### 2026-11-02 11:05 部署到 192.168.2.22（43 提交 + 删夜猫子上线）

- 服务器 `/home/appsrv/workbuddy2api-panel`：`4b5f304 → a6ae90b`（`git pull --ff-only`，本次 GitHub 拉取未失败，
  没走 bundle 通道）。
- 镜像回滚点：`workbuddy2api-panel-wb2api:pre-20261102-1101`（上一版 `:pre-20261030-1040` 仍在）。
- 重建：`docker compose up -d --build` → 容器 `workbuddy2api` Recreated / Started，`Up (healthy)`，**RestartCount=0**。
- 验证：
  - `/healthz` → `{"healthy":2,"realm_servable":{"cn":true,"global":false},"service":"workbuddy2api","total":2}`；`/panel/` HTTP 200。
  - 启动日志出现 **`[upstream] TLS 指纹复刻已启用（JA3 71dc8c533dd919ae9f4963224a4ba8fd / JA4 t13d0a1200_5d04281c6031_b76c5ee2507d）`**，
    即 JA3+JA4、头顺序、Connection 头这组对齐真正生效；`夜猫子已启用` 那行**已消失**，排程只剩签到 / 旅行 / 活跃 / 保活 / 成长 + 补领清扫。
  - `WARN|ERROR` **0 条**；运行 5 分钟内 **7 条请求全部 200、全 `out=success`**，非 200 零条。
  - 其中 `#004 | cn:fast-model | sync | 200 | tok=1 | total=1.9s | out=success` 是我们主动用 curl 打的一次真实对话，
    走的就是新的 TLS 指纹客户端，端到端通；其余 6 条是 `src=192.168.2.127` 的正常客户端流量。
  - 面板 `app.js`：`代理池` 命中 2、`请求上下文` 命中 2、`realm 级默认出口（issue #39）` 在；`夜猫子` / `blackcat` **命中 0**。
  - 数据恢复：44 个用量桶 / 333 条积分流水 / 3 个余额基线，未见异常。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261102-1101 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或服务器 `git reset --hard 4b5f304` 后重建。
- 注意：生产 `config.json` 里 `blackcat_enabled` / `blackcat_hours` / `night_account_cap` 三个键仍留着，
  JSON 未知名会被忽略，面板保存一次配置即自动清掉；不影响运行。

### 2026-10-30 10:48 修「发现新版本」误报并重新部署

- 症状：面板长期提示「发现新版本 v2026.10.29.2131」。线上 `/panel/api/update` 回
  `current_version=dev`、`update_available=true`、`can_install=false`。
- 根因两条：① `Dockerfile` 的 `ARG APP_VERSION=dev`，compose 没传 build arg，二进制 `appVersion=dev`；
  ② `updater.Check` 里 `UpdateAvailable = !currentOK || newest > current`，把「版本号解析失败」（dev）
  直接当成「一定比 release 旧」→ 上游每打一个 tag 就误报一次，而 dev 构建永远没有可比对的版本。
- 改法（提交 `110d0e6`，前缀 `local:`）：`Status` 增加 `local_build`；`Check` 里
  `UpdateAvailable = currentOK && newest > current`，解析失败只报查到的最新 release；
  `canInstallLocked` 复用新增的 `isTimestampedVersion()`；前端 `local_build` 显示
  「本地构建（未注入版本号）· 跳过版本检查」，toast 同步。新增
  `TestCheckLocalBuildDoesNotReportUpdate` + 前端断言。
- 校验：`go build` / `node --check` / `go test ./...` 全绿；线上复查
  `local_build=true`、`update_available=false`、`can_install=false`，面板文案已换。
- 部署：服务器 github 拉取连续失败（`GnuTLS recv error` / `curl 28` 超时）→ 改用
  `git bundle` + `scp` 走文件通道：本地 `git bundle create c39b589..local` →
  `scp /tmp/wb2api-110d0e6.bundle` → 服务器 `git fetch <bundle> local:sync-incoming` →
  `git merge --ff-only sync-incoming`，**以后再遇服务器拉不动可照这条走**。
- 镜像回滚点：`workbuddy2api-panel-wb2api:pre-20261030-1040`（上一版 `:pre-20261030-1029` 仍在）。
- 部署后状态：容器 `Up (healthy)`、`/panel/` 200、请求 7 条全 200、`WARN|ERROR` 0 条。
- ⚠️ 观察到账号数从 3 掉到 2：`auths/` 里 `892aa042`（zhipeng970218@gmail.com）的
  `.json` 与 `.json.bak` 在 **10:36** 一起消失（目录 mtime 记录），两个容器启动点之间；
  两次部署只重建镜像、不碰 `auths/`，`admin_audit.log` 因 `admin.enabled=false` 无记录，
  疑似面板上手动删除。需要该号时重新登录即可。

### 2026-10-30 10:13 跟进上游 9 个提交（`v2026.10.29.2131` 之后），跳过删除类 1 条

- 巡检：`main` 落后 14 个提交，`inspect.sh` 自动 ff `9232c89 → 5373578`（未产生提交）；
  `local` @ `eaf09e7`，工作区干净，上游可达（8 heads）。
- 回滚 tag：**`pre-sync-20261030-101335`**（= `eaf09e7`，`git reset --hard pre-sync-20261030-101335` 可整段退回）。
- 搬入（`cherry-pick -x -n` + 重写为 `port:` 前缀，保留 `(cherry picked from commit …)` 溯源，**9 条全部零冲突**）：
  - `8403247` port: fix: preserve routing policy and validate embedding responses ← `5ad7b35`
  - `69e697e` port: feat(panel): 账号池支持「国内版 / 国际版」分类切换与标签自适应布局修复 ← `6cdc6fb`
  - `2344e13` port: fix: isolate chat soft rate limits by model and normalize CN Kimi tools ← `e8eca81`
  - `ea7ee7c` port: test(panel): 账号池渲染桩补 accRealmFilter/lastAccountsList ← `4850b45`
  - `2ccfa0d` port: fix(embeddings): 6020/6021 改 429 + Retry-After，并与 chat 同口径带上等待时长 ← `6d595fa`
  - `d737c78` port: fix(upstream): Kimi 工具兼容只展开 $ref，去掉按工具名硬编码裁剪 ← `854e74c`
  - `bd2fd76` port: fix(server): 非流式「调用方断开」不再记成 502 upstream_parse，改记 499 ← `7d43b96`
  - `329e8d0` port: fix(server): 号都在途占满时等一个名额，不再直接 503（issue #117）← `9f70e00`
  - `211ad1c` port: fix(panel): 账号任务列表里没有的 code 不再白打上游（下架任务的空转）← `046b1fd`
- 跳过：`96a33f1`（删 4 个下架任务 + 夜猫子排程）、`5373578`（依赖上一条的纯注释）。
  3 个 PR merge（`a5e8e9c`/`22868bb`/`fd91f42`）内容分别等同 `5ad7b35`/`e8eca81`/`6cdc6fb`，不重复搬。
- 收窄判定：**零收窄**。`5ad7b35` 在降级链加的两道护栏（`allowedModel` / `hiddenModel`）拦的是
  「降级到白名单外或隐藏模型」，属修 bug；`e8eca81` 把普通 429 从整号冷却改为（账号+模型）冷却，
  是放宽不是收窄；`854e74c` 去掉 `mcp__rea__` / `mcp__tencent_docs__` 前缀裁剪同样是放宽。
- 校验：`go build ./...`、`go vet ./...` 通过；`go test ./...` **26 包全绿**（panel 31s / scheduler 22s / server 21s / updater 11s）；
  面板路由 **54** / 服务端路由 **17** / `NormalizeAllowedModels` **12** / `autoActions` 条目 **37**（= 同步前，未随 `96a33f1` 减少）；
  `git grep misakano7545` 仅命中 module 路径、import 与 `appRepository` / updater 仓库常量（后两者同步前既在，非本次引入）。
- 残留差异：`git diff main..local -- '*.go' '*.js' '*.html'` 恰好等于 `96a33f1` 的文件集
  （21 个代码文件，+565/−105），即「刻意保留的删除反面」，不多不少。
- 未推远端：`origin/local` 尚停在 `eaf09e7`，需要时 `git push origin local`。
- 待办（下次同步时一起做，避免在同步前弄脏工作区）：`.fork-sync-baseline` 是 **CRLF** 行尾，
  `inspect.sh` 第 6 段用 `read -r` 读出的 `path` / 正则各带一个 `\r`，三项计数全部输出 `0`（假 0）；
  本次靠手工 grep 核对（数值与基线一致）。修法二选一：`sed -i 's/\r$//' .fork-sync-baseline`，
  或让脚本在 `read` 后剥掉 `\r`（更耐造，别人用 Windows 编辑器改这文件不会再中招）。

### 2026-10-30 10:33 部署到 192.168.2.22（上游 9 提交上线）

- 服务器 `/home/appsrv/workbuddy2api-panel`：`eaf09e7 → 465960c`（`git pull --ff-only`，27 文件 +1162/−45，无冲突）。
- 镜像回滚点：`docker tag workbuddy2api-panel-wb2api:latest ...:pre-20261030-1029`（旧镜像 `09fc921` 保留）。
- 重建：`docker compose up -d --build` → 容器 `workbuddy2api` Recreated / Started，`Up (healthy)`。
- 验证：`/healthz` → `{"healthy":3,"realm_servable":{"cn":true,"global":true},"total":3}`；
  `/panel/` 200；`app.js` 中 `openRemarks` 命中 3（备注入口在）；近 5 分钟 `WARN|ERROR` **0 条**；请求日志 3 条全 200。
- 顺带实证「夜猫子」：启动日志打出「夜猫子已启用：[23] 点」，坐实 `96a33f1` 不能跟（见上表）。
- 回滚：`docker tag workbuddy2api-panel-wb2api:pre-20261030-1029 workbuddy2api-panel-wb2api:latest`
  + `docker compose up -d --force-recreate`；或 `git reset --hard pre-sync-20261030-101335` 后重建。
- 服务器访问方式（记一笔，别再走密码）：本机密钥 `~/.ssh/id_wb_222`，`ssh -i ~/.ssh/id_wb_222 root@192.168.2.22`。
  `app` 账户无密钥、密码认证也不通。

### 2026-10-29 15:41 服务器仓库规范化：`origin` 改指本 fork，日常同步只剩 `git pull`

- 服务器 `/home/appsrv/workbuddy2api-panel` 原先 `origin` = **上游**（clone 时就留下的），
  我们的 fork 只能另起名 `fork`，每次同步得写 `git fetch fork local && git merge --ff-only fork/local`。
- 改为：`git remote rename origin upstream && git remote rename fork origin`。
  现在与本地仓库命名一致：`origin` = 本 fork，`upstream` = 上游；
  `local` 跟踪 `origin/local`，`main` 跟踪 `upstream/main`。
- 日常更新流程因此收敛为两条：
  ```bash
  git pull --ff-only          # 服务器：只同步代码/文档
  docker compose up -d --build   # 只有动到 Go / 前端 / Dockerfile 才需要
  ```
- 实测：`git pull --ff-only` 把 `2bdac00 → 6849152` 拉平（此前 GitHub TLS 抖动失败的补拉），
  容器仍 `Up (healthy)`；顺带 `git fetch upstream main:main` 把服务器的 `main` 也快进到 `9232c89`。
- 服务器 `main` 不参与部署，只作为上游镜像存在；部署永远由 `local` 构建。

### 2026-10-29 15:33 fork 的 `main` 快进到上游并推送

- 现象：fork 页面上 `main` 一直停在 `bb6d581`，看起来「上游没更新」。
- 实际：上游今天推了 6 个提交（`4d643b1` 07:13 → `9232c89` 14:16），其中 `a25192e`、`9232c89`
  是下午 14 点前后才进的；fork 的 `main` 只是本地镜像分支，**不会自动跟进**。
- 本地 `main` 在 15:17 已被巡检脚本快进到 `9232c89`，但**从没 push 过**，所以远端仍是 `bb6d581`。
- 本次：`git switch main && git merge --ff-only upstream/main && git push origin main`
  → `bb6d581..9232c89`（纯 ff，未产生提交）。
- 现状：`main` = `upstream/main` = `origin/main` = **9232c89**；`local` = `2bdac00`（内容等价，见上一条）。
- 提醒：fork 页面上 `main` 显示的日期是「最后一次推送时间」，不是上游最后一次提交时间——
  判断上游有没有更新要 `git fetch upstream` 后看 `upstream/main`，别看 fork 页面。

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
