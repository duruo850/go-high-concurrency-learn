# 8. 链路追踪：OpenTelemetry → Jaeger

## 目标

跑通 **应用埋点 → OTLP 导出 → Jaeger 存储/查询**，看清一次请求里各 span 的耗时瀑布，并理解它和 Prometheus / ELK / Sentry 的分工。

## 前置

- Docker Desktop
- 完成 [1plan-basics.md](./1plan-basics.md)

## 框架（必记）

```text
业务代码 / HTTP 中间件 创建 Span
        → OTLP（HTTP :4318）导出
        → Jaeger（Collector + 存储 + UI）
        → 按 TraceID 查看整条调用链
```

| 概念 | 一句话 |
|------|--------|
| **Trace** | 一次完整请求的因果链 |
| **Span** | 链上的一段工作（HTTP / DB / RPC…） |
| **TraceID** | 全链路同一 ID；可写进日志做关联 |
| **OTLP** | OpenTelemetry 标准导出协议 |

| 和谁比 | 差异 |
|--------|------|
| vs Prometheus | 指标是聚合数字；追踪是**单次请求路径与耗时** |
| vs ELK | 日志是事件文本；追踪是**结构化父子时间线** |
| vs Sentry | Sentry 偏错误聚合；追踪偏**正常路径延迟拆解**（也可标 Error span） |

本仓库：**OTel Go SDK + otelhttp Filter + `/api/trace-demo` 子 span → Jaeger all-in-one**。

## 步骤

1. 启动 Jaeger

```bash
make jaeger-up
```

UI：http://127.0.0.1:16686 ；OTLP HTTP：`http://127.0.0.1:4318`

2. 打开追踪并启动服务

任选其一：

- 环境变量（推荐实验）：

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
make run
```

- 或改 `configs/config.yaml`：

```yaml
trace:
  endpoint: "http://127.0.0.1:4318"
  sample_ratio: 1.0
```

期望启动日志含：`trace: enabled → ...`

3. 打一条带父子 span 的请求

```bash
make trace-demo
# 等价：curl -sS http://127.0.0.1:8000/api/trace-demo
```

响应里有 `trace_id` 和 `steps`（auth → db → downstream）。

4. 在 Jaeger UI 查看

- Service 选 `go-high-concurrency-learn` → **Find Traces**
- 点开一条：应看到 `GET /api/trace-demo`，其下有 `trace.demo.handle` → `auth.validate` / `db.query` / `downstream.rpc`（下游 span 带 Error 演示）
- 也可用响应里的 `trace_id` 在 Search 里按 Trace ID 查

5. （可选）普通接口也会有 HTTP server span

```bash
curl -sS 'http://127.0.0.1:8000/api/echo?msg=traced'
```

日志字段可带 `trace_id`（启用追踪时）。

6. 关闭

```bash
make jaeger-down
# 或取消 endpoint 后重启服务
```

## 验收

- [ ] 能口述：Trace / Span / TraceID / OTLP 各干什么
- [ ] `make jaeger-up` + 启用 endpoint 后，`make trace-demo` 能在 Jaeger 看到瀑布
- [ ] 能区分：追踪看「这条请求慢在哪」；Prometheus 看「整体是否变慢」；ELK 看「日志原文」；Sentry 看「异常 Issue」

## 实现落点（本仓库）

| 位置 | 作用 |
|------|------|
| `internal/observability/tracex/` | Init / Shutdown / Filter / DemoWalk |
| `internal/server/http` | `khttp.Filter(tracex.HTTPFilter)` + `/api/trace-demo` |
| `configs/config.yaml` → `trace` | endpoint / sample_ratio |
| `docker-compose` profile `trace` | Jaeger all-in-one |
| `make jaeger-up` / `make trace-demo` | 一键实验 |

## 下一篇

→ [7plan-compare.md](./7plan-compare.md)（把 Tracing 一并对照）
