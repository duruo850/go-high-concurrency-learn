# 1. 基础：跑通 Demo 服务

## 目标

理解本仓库是一个「故意可泄漏」的 Kratos HTTP 服务，先证明业务正常，再学观测工具。

## 前置

- Go 1.23+
- 工作目录：`go-memleak-demo/`

## 步骤

1. 安装依赖并测试

```bash
go mod tidy
make test
```

期望：全部通过。

2. 读配置 `configs/config.yaml`

- `leak.mode: leak | fixed`：泄漏或对照
- `server.http.addr`：默认 `:8000`
- `log.path`：JSON 日志文件（给 ELK 用）
- `sentry.dsn`：空则不上报（给 Sentry 用）

3. 启动

```bash
make run
```

期望日志含 `leak: starting mode=leak` 与 `listening on: ...:8000`。

4. 业务自检

```bash
make echo
```

期望 `/health` → `{"status":"ok"}`，`/api/echo?msg=hello` → `{"message":"hello"}`。

5. （可选）改成 `leak.mode: fixed` 重启，感受「不泄漏」对照；再改回 `leak`。

## 验收

- [ ] `make test` 绿
- [ ] `make echo` 三个请求均 HTTP 200
- [ ] 能说清：泄漏在启动时自动开始，没有 `/leak/*` 触发接口

## 下一篇

→ [2plan-delve.md](./2plan-delve.md)
