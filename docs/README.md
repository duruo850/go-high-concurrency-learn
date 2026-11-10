# 可观测性学习路径（按顺序）

每个文件是一个独立实验计划：先读计划 → 按步骤动手 → 打勾验收。实现代码都在本仓库，命令相对仓库根 `go-memleak-demo/`。

| 顺序 | 计划 | 学什么 | 动手入口 |
|------|------|--------|----------|
| 1 | [1plan-basics.md](./1plan-basics.md) | 跑通服务、配置、业务接口 | `make test` / `make run` / `make echo` |
| 2 | [2plan-delve.md](./2plan-delve.md) | Delve：协程卡住在哪 | `dlv debug ...` |
| 3 | [3plan-pprof.md](./3plan-pprof.md) | pprof：堆涨归因到函数 | `make pprof-capture` / `make pprof-top` |
| 4 | [4plan-prometheus.md](./4plan-prometheus.md) | Prometheus：指标监控框架 | `make prom-up` + Graph |
| 5 | [5plan-elk.md](./5plan-elk.md) | ELK：日志采集与检索 | `make elk-up` + Kibana |
| 6 | [6plan-sentry.md](./6plan-sentry.md) | Sentry：异常事件框架 | 配 DSN + `make sentry-boom` |
| 7 | [7plan-compare.md](./7plan-compare.md) | Prometheus / ELK / Sentry 对照 | 复盘选型 |
| 8 | [knowledge/8plan-million-short-conn.md](./knowledge/8plan-million-short-conn.md) | 架构 + 单进程在线量级 + 百万短连方案 | 先读文首 §A/§B，再读消峰与限流 |

**分工速记**

```text
Prometheus  →  CPU / 内存 / goroutine 等「数字趋势」
ELK         →  日志全文检索，排查「发生了什么」
Sentry      →  未处理异常 / 错误聚合成 Issue，「哪段代码炸了」
Delve/pprof →  本机调试与堆定因（偏开发工具，不是监控平台）
```

## 专题知识

| 文档 | 内容 |
|------|------|
| [knowledge/8plan-million-short-conn.md](./knowledge/8plan-million-short-conn.md) | 百万短连接 / 高并发服务端架构（含 SVG） |
| [knowledge/av-signaling-sfu-mcu-concurrency.md](./knowledge/av-signaling-sfu-mcu-concurrency.md) | 音视频传输：信令 + SFU/MCU，以及音视频高并发架构 |
