# 可观测性学习路径（按顺序）

每个文件是一个独立实验计划：先读计划 → 按步骤动手 → 打勾验收。实现代码都在本仓库，命令相对仓库根 `go-high-concurrency-learn/`。

| 顺序 | 计划 | 学什么 | 动手入口 |
|------|------|--------|----------|
| 1 | [plan/1plan-basics.md](./plan/1plan-basics.md) | 跑通服务、配置、业务接口 | `make test` / `make run` / `make echo` |
| 2 | [plan/2plan-delve.md](./plan/2plan-delve.md) | Delve：协程卡住在哪 | `dlv debug ...` |
| 3 | [plan/3plan-pprof.md](./plan/3plan-pprof.md) | pprof：堆涨归因到函数 | `make pprof-capture` / `make pprof-top` |
| 4 | [plan/4plan-prometheus.md](./plan/4plan-prometheus.md) | Prometheus：指标监控框架 | `make prom-up` + Graph |
| 5 | [plan/5plan-elk.md](./plan/5plan-elk.md) | ELK：日志采集与检索 | `make elk-up` + Kibana |
| 6 | [plan/6plan-sentry.md](./plan/6plan-sentry.md) | Sentry：异常事件框架 | 配 DSN + `make sentry-boom` |
| 7 | [plan/8plan-tracing.md](./plan/8plan-tracing.md) | 链路追踪：OTel → Jaeger | `make jaeger-up` + `make trace-demo` |
| 8 | [plan/7plan-compare.md](./plan/7plan-compare.md) | Prometheus / ELK / Sentry / Tracing 对照 | 复盘选型 |

**分工速记**

```text
Prometheus  →  CPU / 内存 / goroutine 等「数字趋势」
ELK         →  日志全文检索，排查「发生了什么」
Sentry      →  未处理异常 / 错误聚合成 Issue，「哪段代码炸了」
Tracing     →  单次请求跨 span 耗时瀑布，「慢在哪一段」
Delve/pprof →  本机调试与堆定因（偏开发工具，不是监控平台）
```

## 专题知识

| 文档 | 内容 |
|------|------|
| [knowledge/1plan-million-short-conn.md](./knowledge/1plan-million-short-conn.md) | 百万短连接 / 高并发服务端架构 |
| [knowledge/av-signaling-sfu-mcu-concurrency.md](./knowledge/av-signaling-sfu-mcu-concurrency.md) | 音视频传输：信令 + SFU/MCU，以及音视频高并发架构 |
| [knowledge/high-concurrency-full-arch.svg](./knowledge/high-concurrency-full-arch.svg) / [knowledge/high-concurrency-arch.svg](./knowledge/high-concurrency-arch.svg) | 架构图 |
