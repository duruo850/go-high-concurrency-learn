# 5. ELK：日志框架（排查「发生了什么」）

## 目标

跑通 **应用写 JSON 日志 → Filebeat → Elasticsearch → Kibana 检索**，理解日志链路与 Prometheus 的差异。

## 前置

- Docker Desktop（ES 建议本机内存 ≥ 4GB 可用）
- 完成 [1plan-basics.md](./1plan-basics.md)
- `configs/config.yaml` 中 `log.path` 默认为 `tmp/logs/app.jsonl`

## 框架（必记）

```text
应用 / Agent 采集 → Filebeat/Logstash 转发 → Elasticsearch 存储检索 → Kibana 可视化
```

本仓库：**Filebeat + Elasticsearch + Kibana**（未强制上 Logstash，学习链路足够）。

## 步骤

1. 确保日志目录存在并启动服务（会写 JSON 行日志）

```bash
mkdir -p tmp/logs
make run
```

2. 造几条业务日志

```bash
make elk-sample
```

会请求 `/api/echo` 与 `/api/log-demo`，向 `tmp/logs/app.jsonl` 追加 JSON。

3. 启动 ELK

```bash
make elk-up
```

等待约 30～60s，打开 Kibana：http://127.0.0.1:5601

4. Kibana 操作（首次）

- Stack Management → Data Views（或 Index Patterns）
- 创建匹配 `go-memleak-demo-*`（或 Filebeat 实际 index）
- Discover 里搜 `go-memleak-demo` / `log_demo` / `echo`

5. 对照文件内容

```bash
tail -n 5 tmp/logs/app.jsonl
```

6. 关闭

```bash
make elk-down
```

## 期望

- `app.jsonl` 每行一个 JSON（含 `ts`、`level`、`msg`、`service` 等）
- Kibana Discover 能搜到刚才的 `log_demo` 事件

## 验收

- [ ] 能画出 ELK 四环，并说明本 demo 用 Filebeat 而不是 Logstash
- [ ] 能说清：ELK 查日志现场；堆泄漏定因仍靠 pprof
- [ ] Prometheus=指标 Pull；ELK=日志 Push

## 下一篇

→ [6plan-sentry.md](./6plan-sentry.md)
