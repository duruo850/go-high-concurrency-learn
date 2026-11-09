# Go Goroutine / 堆内存泄漏 Demo

Kratos HTTP 演示服务：启动即按配置泄漏（或 fixed 对照），业务接口正常；配套 Prometheus / pprof / Delve 可复制调试剧本。

## 前置条件

- Go 1.23+（首次 `go` 命令可能会自动下载匹配的 toolchain）
- Docker Desktop（Prometheus）
- Delve：`go install github.com/go-delve/delve/cmd/dlv@latest`

工作目录一律为仓库根：`go-memleak-demo/`。

```bash
cd go-memleak-demo
go mod tidy
```

## 配置

`configs/config.yaml`：

| 字段 | 说明 |
|------|------|
| `leak.mode` | `leak`：永不释放；`fixed`：分配后丢弃，供对照 |
| `leak.workers` | 后台 worker 数（默认 8） |
| `leak.chunk_bytes` | 每次分配字节数 |
| `leak.interval` | 分配间隔（如 `200ms`） |
| `server.http.addr` | 默认 `0.0.0.0:8000` |

泄漏在进程启动时自动开始，**没有** `/leak/*` 触发接口。

## Makefile 一览

| 目标 | 作用 |
|------|------|
| `make test` | 单元 / httptest 功能测试 |
| `make run` | 启动服务（自动泄漏） |
| `make echo` | curl `/health`、`/api/echo` |
| `make metrics-watch` | 循环打印 `go_goroutines` / heap |
| `make prom-up` / `make prom-down` | Docker Prometheus |
| `make pprof-capture` | 间隔抓 `tmp/pprof/h1.out`、`h2.out` |
| `make pprof-top` | 终端 diff top / list（定因） |
| `make pprof-diff` | 浏览器 pprof UI（diff） |

---

## A. 功能验收

**前置：** 已 `go mod tidy`。

```bash
make test
```

期望：全部 `ok` / `[no test files]`，无 `FAIL`。

另开终端启动服务：

```bash
make run
```

期望日志含：

- `leak: starting mode=leak workers=8 ...`
- `[HTTP] server listening on: [::]:8000`

再开终端：

```bash
make echo
```

期望样例：

```text
==> GET /health
{"status":"ok"}
HTTP 200
==> GET /api/echo?msg=hello
{"message":"hello"}
HTTP 200
==> GET /api/echo (default pong)
{"message":"pong"}
HTTP 200
```

---

## B. 用 metrics 看见上涨（无需外部工具）

**前置：** `make run` 已在跑。

循环观察：

```bash
make metrics-watch
```

或两次对比（约 30s）：

```bash
curl -sS http://127.0.0.1:8000/metrics | grep -E '^go_goroutines |^go_memstats_heap_alloc_bytes '
sleep 30
curl -sS http://127.0.0.1:8000/metrics | grep -E '^go_goroutines |^go_memstats_heap_alloc_bytes '
```

本机实测样例（约 15s 间隔）：

```text
# sample 1
go_goroutines 20
go_memstats_heap_alloc_bytes 1.05201864e+08

# sample 2
go_goroutines 18
go_memstats_heap_alloc_bytes 1.45132648e+08
```

期望：

- `go_memstats_heap_alloc_bytes` **明显增大**
- `go_goroutines` 稳定高于 workers 基线（约 8 个 `leakWorker` + HTTP/runtime）

也可用：`./scripts/metrics-once.sh`

---

## C. Delve 真实会话

**前置：** 停掉占用 `:8000` 的 `make run`（否则端口冲突）。安装最新 Delve。

1. 安装

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

若提示 Go 版本过新：加 `--check-go-version=false`。

2. 启动调试（项目根）

```bash
dlv debug ./cmd/server --check-go-version=false -- -conf configs
```

3. 打断点（**真实符号名**，本仓库已验证）

```text
break github.com/TripleCGame/go-memleak-demo/internal/service/leak.(*Leaker).leakWorker
```

期望：`Breakpoint 1 set at ... leak.go:93`

4. 继续执行

```text
continue
```

期望命中：

```text
> [Breakpoint 1] ... (*Leaker).leakWorker() ./internal/service/leak/leak.go:93
```

5. 看协程 / 栈 / 参数

```text
goroutines
stack
args
```

本机实测关键点：

- `goroutines` 中可见多个卡在 `leak.go` 的 worker（命中的那一个在 `leakWorker`，其余可能还在 `Start.func1`）
- `stack` 顶部为 `(*Leaker).leakWorker` ← `Start.func1`
- `args` 可见 `id`（worker 编号）与 `l`（`*Leaker`）

说明：Delve 证明「协程卡在分配循环」；堆归因下一步用 pprof。

退出：`quit` → `y`

---

## Prometheus 观测框架

标准链路：

```text
Exporter 暴露指标 → Prometheus 拉取并存储 → Alertmanager 处理告警 → Grafana 可视化
```

| 环节 | 职责 | 本 demo 对应 |
|------|------|----------------|
| **Exporter** | 进程（或 sidecar）以 Prometheus 文本格式暴露 `/metrics` | 本服务即 Exporter：`GET :8000/metrics`（`promhttp` + Go runtime 指标） |
| **Prometheus** | 按 `scrape_configs` **主动拉取**（pull），写入 TSDB，提供 PromQL | `make prom-up` → `:9090`；配置见 `deploy/prometheus.yml` |
| **Alertmanager** | 接收 Prometheus 触发的告警，做分组 / 抑制 / 路由（邮件、IM 等） | **未部署**；生产中由 Prometheus `alerting` + 规则文件对接 |
| **Grafana** | 以 Prometheus 为数据源做仪表盘与可视化 | **未部署**；本 demo 先用 Prometheus 自带 Graph；生产再接 Grafana |

数据流示意：

```text
┌─────────────────────┐  scrape /metrics   ┌──────────────────┐  告警事件   ┌──────────────┐
│  Exporter（本服务）  │ ─────────────────► │   Prometheus     │ ─────────► │ Alertmanager │
│  :8000/metrics      │   pull + 存储 TSDB  │   :9090          │            │  分组/通知    │
└─────────────────────┘                    └────────┬─────────┘            └──────────────┘
                                                    │ PromQL 查询
                                                    ▼
                                           ┌──────────────────┐
                                           │     Grafana      │
                                           │   仪表盘可视化    │
                                           └──────────────────┘
```

要点：

1. **Pull 模型**：指标由 Exporter 暴露，Prometheus 定时来拉，而不是业务主动推（Pushgateway 是例外场景）。
2. **本仓库已跑通前两环**：Exporter（服务）→ Prometheus（Docker）；Alertmanager / Grafana 属完整栈后续扩展。
3. 泄漏演示可先用 `/metrics` + Prometheus Graph 看见上涨，再用 pprof / Delve 定因（见 B / D / E）。

---

## D. Docker Prometheus（Exporter → Prometheus）

**前置：** Docker Desktop 已启动；本机 `make run` 在 `:8000` 监听（Exporter 已就绪）。

1. 启动 Prometheus（拉取并存储）

```bash
make prom-up
```

2. 浏览器打开 Targets：<http://127.0.0.1:9090/targets>

期望：`go-memleak-demo` → **UP**（说明 scrape Exporter 成功）。  
macOS 上 scrape 地址为 `host.docker.internal:8000`（`docker-compose.yaml` 已配 `extra_hosts: host.docker.internal:host-gateway`）。

本机 API 核对：

```bash
curl -sS 'http://127.0.0.1:9090/api/v1/targets' | grep -o '"health":"[^"]*"'
```

期望含 `"health":"up"`。

3. Graph 粘贴 PromQL（与 Exporter `/metrics` 指标名一致）：

```promql
go_goroutines
```

```promql
go_memstats_heap_alloc_bytes
```

```promql
rate(go_memstats_alloc_bytes_total[1m])
```

期望：堆相关曲线随时间右上扬；`go_goroutines` 维持在较高水平。

本机查询样例：`go_goroutines ≈ 20`，`go_memstats_heap_alloc_bytes` 为数亿字节且持续增大。

> 完整栈下一步通常是：在 Prometheus 配 recording/alerting rules → Alertmanager；Grafana 添加 Prometheus 数据源做面板。本 demo 到 Graph 查询即可闭环「看见泄漏」。

4. 关闭

```bash
make prom-down
```

---

## E. pprof 定因

**前置：** `make run` 在跑。

1. 抓两次 heap profile（内含约 20s 间隔）

```bash
make pprof-capture
```

期望生成 `tmp/pprof/h1.out`、`tmp/pprof/h2.out`。

2. 终端 diff（推荐先跑这个确认归因）

```bash
make pprof-top
```

请使用 **`-diff_base`**（Makefile 已写好）。本机实测 top 会指向：

```text
github.com/TripleCGame/go-memleak-demo/internal/service/leak.(*Leaker).leakWorker
```

`list` 会落到 `internal/service/leak/leak.go` 中 `buf := make([]byte, l.chunkBytes)` 一行。

单份 profile 也能定因（不对比时）：

```bash
go tool pprof -top http://127.0.0.1:8000/debug/pprof/heap
```

期望 flat 占比接近 99% 落在 `(*Leaker).leakWorker`。

3. 浏览器 UI

```bash
make pprof-diff
```

打开 <http://127.0.0.1:8081>，View → Top / Source，应指向 `leak.go` 的 `leakWorker`。

4. goroutine 文本栈

```bash
curl -sS 'http://127.0.0.1:8000/debug/pprof/goroutine?debug=1' | head -n 30
```

期望看到：

```text
github.com/TripleCGame/go-memleak-demo/internal/service/leak.(*Leaker).leakWorker
```

5. fixed 对照

把 `configs/config.yaml` 中 `leak.mode` 改为 `fixed`，重启 `make run`，再 `make pprof-capture` + `make pprof-top`：  
堆增长应明显变缓（不再把 slice 挂到 `Leaker.held`）；进程退出后可被 GC 回收。演示完改回 `leak`。

---

## 目录结构

```text
cmd/server/                 # 入口：加载配置、启动泄漏、跑 Kratos App
internal/conf/              # 配置结构
internal/service/echo/      # GET /api/echo
internal/service/leak/      # leakWorker / fixedWorker
internal/server/http/       # /health /api/echo /metrics /debug/pprof
configs/config.yaml
deploy/prometheus.yml
docker-compose.yaml
Makefile
scripts/metrics-once.sh
```

## 端口

| 地址 | 用途 |
|------|------|
| `:8000` | 业务 + metrics + pprof |
| `:9090` | Prometheus UI |
| `:8081` | `make pprof-diff` 本地 UI |
