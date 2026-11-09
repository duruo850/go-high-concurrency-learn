# 4. Prometheus：指标监控框架

## 目标

跑通 **Exporter → Prometheus（拉取存储）**，用指标看见内存 / goroutine 上涨。Alertmanager / Grafana 知框架即可，本篇不强制部署。

## 前置

- Docker Desktop
- `make run`（本服务即 Exporter，暴露 `:8000/metrics`）

## 框架（必记）

```text
Exporter 暴露指标 → Prometheus 拉取并存储 → Alertmanager 处理告警 → Grafana 可视化
```

本仓库：**已实现前两环**；后两环见主 README「Prometheus 观测框架」。

## 步骤

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

打开 http://127.0.0.1:9090/targets → `go-memleak-demo` 为 **UP**。

3. Graph 查询

```promql
go_goroutines
go_memstats_heap_alloc_bytes
rate(go_memstats_alloc_bytes_total[1m])
```

4. 关闭

```bash
make prom-down
```

## 验收

- [ ] 能解释 Pull：Prometheus 来拉 `/metrics`，不是业务推指标
- [ ] Targets UP，PromQL 有数据且堆曲线上扬
- [ ] 能区分：Prometheus「看见涨了」≠ pprof「定因到函数」

## 下一篇

→ [5plan-elk.md](./5plan-elk.md)
