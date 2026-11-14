# 6. Sentry：异常事件框架

## 目标

理解 Sentry 处理的是 **错误事件 / Issue**，不是全量日志，也不是 CPU 曲线。本仓库用官方 Go SDK 上报；本地用 **Sentry 免费云账号 DSN**（自建 Sentry 过重，学习不推荐先上）。

## 前置

- 完成 [1plan-basics.md](./1plan-basics.md)
- 注册 https://sentry.io → 创建 Go 项目 → 复制 DSN

## 框架（必记）

```text
SDK 捕获异常/消息 → 上报 Sentry 后端 → 指纹聚合成 Issue → 告警 / 指派 / 关联 Release
```

| 和谁比 | 差异 |
|--------|------|
| vs Prometheus | 不看内存曲线，看「炸了没有、炸几次」 |
| vs ELK | 不做全文日志湖；聚焦 exception + 上下文，自动归并同类错误 |

## 步骤

1. 配置 DSN（任选其一）

- 写进 `configs/config.yaml`：

```yaml
sentry:
  dsn: "https://xxxxx@oxxxx.ingest.sentry.io/xxxxxx"
  environment: local
  debug: true
```

- 或环境变量（优先）：`export SENTRY_DSN='https://...'`

DSN 为空时：服务仍可启动，`/api/boom` 只会本地打日志，**不上报**。

2. 启动

```bash
make run
```

期望启动日志含 `sentry: enabled`（若 DSN 有效）。

3. 触发事件

```bash
make sentry-boom
# 等价：curl -sS http://127.0.0.1:8000/api/boom
```

期望：HTTP 500；Sentry 项目 Issues 出现新 issue（含 stack）。

4. （可选）主动面包屑 / 消息

```bash
curl -sS 'http://127.0.0.1:8000/api/sentry-message?msg=hello-sentry'
```

在 Sentry 看 Message 类事件。

5. 在 Sentry UI 观察

- Issue 标题、堆栈、`environment=local`
- 连续触发多次：应归并到同一 Issue（fingerprint），count 增加

## 验收

- [ ] 能口述 Sentry 三问：哪段代码、多少次、哪个版本/环境
- [ ] 配 DSN 后 `make sentry-boom` 能在控制台看到 Issue
- [ ] 能区分：Sentry≠ELK≠Prometheus

## 下一篇

→ [7plan-tracing.md](./7plan-tracing.md)
