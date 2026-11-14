# 4. Prometheus：指标监控框架

## 目标

弄清 **完整架构各层干什么**，并跑通本仓库已实现的前两环：**Exporter → Prometheus**，用指标看见内存 / goroutine 上涨。Alertmanager / Grafana 本篇讲清楚角色，**不强制部署**（查询先用 Prometheus 自带 Graph）。

## 前置

- Docker Desktop
- `make run`（本服务即 Exporter，暴露 `:8000/metrics`）

---

## 1. 完整架构（必记）

```text
┌──────────────────┐   pull /metrics    ┌─────────────────────┐
│ Exporter         │ ─────────────────► │ Prometheus Server   │
│ 本服务 :8000     │   定时抓取 + 存盘   │ :9090               │
│ promhttp 文本    │                    │ · scrape            │
└──────────────────┘                    │ · TSDB 本地存储      │
                                        │ · PromQL / Graph    │
                                        │ · 评估告警规则       │
                                        └──────────┬──────────┘
                              告警事件              │ 查询（PromQL）
                                                   │
                     ┌─────────────────┐           │
                     │ Alertmanager    │ ◄─────────┤
                     │ 分组/抑制/路由   │           │
                     │ 邮件 / IM / Webhook          │
                     └─────────────────┘           │
                                                   ▼
                                        ┌─────────────────────┐
                                        │ Grafana（推荐生产）  │
                                        │ Dashboard / 告警面板 │
                                        │ 数据源 = Prometheus  │
                                        └─────────────────────┘
```

本仓库现状：

| 环节 | 状态 |
|------|------|
| Exporter | **已有**（`GET :8000/metrics`） |
| Prometheus | **已有**（`make prom-up`，配置 `deploy/prometheus.yml`） |
| Alertmanager | **未部署**（知角色即可） |
| Grafana | **未部署**；学习用 Prometheus UI → Graph；生产再接 Grafana |

---

## 2. 各组件干什么

| 组件 | 干什么 | 不是什么 |
|------|--------|----------|
| **Exporter** | 以 Prometheus 文本格式暴露当前指标快照（`/metrics`） | 不负责长期存历史；不负责画图 |
| **Prometheus Server** | **主动 Pull** 各 target；写入本地 **TSDB**；提供 PromQL、简易 Graph、规则评估 | 不是日志系统；不是全量请求明细 |
| **Alertmanager** | 接收 Prometheus 发出的告警，做分组、抑制、静默、路由到通知渠道 | 自己不算指标；规则仍写在 Prometheus 侧 |
| **Grafana** | 以 Prometheus 为 **数据源**，做仪表盘、多数据源对照、更丰富告警/分享 | 不替代 Prometheus 存储；PromQL 仍在 Prometheus 上算 |

### Exporter 用的 Go 库（本仓库）

| | |
|--|--|
| **库** | [`github.com/prometheus/client_golang`](https://github.com/prometheus/client_golang) |
| **HTTP 暴露** | `prometheus/promhttp` → `promhttp.Handler()` |
| **默认指标** | 挂上 Handler 即带 **Go runtime + process** collector（`go_*` / `process_*`） |
| **自定义指标** | 同库的 `prometheus.NewCounter/Gauge/Histogram` + `MustRegister`（本 demo 未加业务指标） |
| **本仓库落点** | `internal/server/http/http.go`：`srv.Handle("/metrics", promhttp.Handler())` |
| **go.mod** | `github.com/prometheus/client_golang v1.20.5` |

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// 一行即可当 Exporter：Prometheus 来拉这个路径
srv.Handle("/metrics", promhttp.Handler())
```

**Pull 模型：** 业务进程只暴露 `/metrics`，由 Prometheus 按 `scrape_interval` 来拉。  
例外才是 Pushgateway（短任务 / 批处理），本仓库不用。

本仓库 scrape（`deploy/prometheus.yml`）：每 **5s** 拉 `host.docker.internal:8000/metrics`。

---

## 3. 存储方案

| 方案 | 说明 | 本仓库 |
|------|------|--------|
| **本地 TSDB** | Prometheus 进程内置时序库，按块落盘；适合单机 / 中小规模 | **默认就是这个**（Docker 容器内 `/prometheus`） |
| **保留时长** | `--storage.tsdb.retention.time`（如 `15d`）；到期删旧块 | demo 未特调，用镜像默认 |
| **Remote Write** | 把样本异步写到长期存储（如 Thanos Receive、Mimir、Cortex、VictoriaMetrics） | **未接**；生产要跨机房/超长保留再上 |
| **Remote Read / 联邦** | 从远端读或 Prometheus 联邦聚合 | 学习阶段可忽略 |
| **对象存储 + Thanos/Mimir** | 无限期历史、多租户、高可用 | 生产级方案，本 demo 不部署 |

学习结论：**先搞懂「Prometheus 自己带 TSDB」**；Grafana **不存** Prometheus 指标副本（除非另配），只查询 Prometheus。

---

## 4. 可视化：Prometheus Graph vs Grafana

| | **Prometheus Graph**（`:9090` → Graph） | **Grafana** |
|--|------------------------------------------|-------------|
| 用途 | 临时敲 PromQL、看单曲线是否上涨 | 固定 Dashboard、多图布局、分享、值班大盘 |
| 数据从哪来 | 查本机 TSDB | 添加 **Prometheus 数据源**（URL 指向 `:9090`） |
| 本仓库 | **实验用这个就够** | **可选扩展**；架构上要知道「生产看盘通常是 Grafana」 |
| 和告警 | 规则在 Prometheus；通知靠 Alertmanager | Grafana 也可配自己的告警，但常见仍是 Prom 规则 → Alertmanager |

口诀：**Prometheus 负责采 + 存 + 算；Grafana 负责好看的看板；Alertmanager 负责把火发出去。**

---

## 5. 本仓库要盯的指标（泄漏）

详情与完整列表见 [8plan-compare.md](./8plan-compare.md)「时序指标」节。实验最少三条：

```promql
go_goroutines
go_memstats_heap_alloc_bytes
rate(go_memstats_alloc_bytes_total[1m])
```

---

## 6. 步骤

1. 不启 Docker，先直接看 Exporter

```bash
make metrics-watch
# 或
curl -sS http://127.0.0.1:8000/metrics | grep -E '^go_goroutines |^go_memstats_heap_alloc_bytes '
```

隔 30s 再 curl 一次，heap 应增大。

2. 启动 Prometheus

```bash
make prom-up
```

打开 http://127.0.0.1:9090/targets → `go-high-concurrency-learn` 为 **UP**。

3. 用自带 Graph 查询（上节三条 PromQL）

4. （可选，不强制）本地再起 Grafana，数据源填 `http://host.docker.internal:9090`，面板贴同一 PromQL——体会「存仍在 Prometheus，盘在 Grafana」。

5. 关闭

```bash
make prom-down
```

---

## 验收

- [ ] 能画出四环：Exporter → Prometheus → Alertmanager → Grafana，并说出每环职责
- [ ] 知道存储默认是 **本地 TSDB**；Grafana 不替代存指标
- [ ] 能解释 Pull：Prometheus 来拉 `/metrics`，不是业务推指标
- [ ] Targets UP，Graph 有数据且堆曲线上扬
- [ ] 能区分：Prometheus「看见涨了」≠ pprof「定因到函数」

## 下一篇

→ [5plan-elk.md](./5plan-elk.md)
