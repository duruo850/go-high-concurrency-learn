# 2. Delve：调试「协程卡在哪」

## 目标

用调试器证明泄漏 worker 落在 `leakWorker` 循环，而不是猜调用栈。

## 前置

- 完成 [1plan-basics.md](./1plan-basics.md)
- `go install github.com/go-delve/delve/cmd/dlv@latest`
- **停掉**占用 `:8000` 的 `make run`

## 概念

| 工具 | 回答的问题 |
|------|------------|
| Delve | 断点时刻：哪个 goroutine、栈、参数是什么？ |
| pprof（下一篇） | 一段时间后：堆增量归到哪个函数？ |

## 步骤

1. 启动调试

```bash
dlv debug ./cmd/server --check-go-version=false -- -conf configs
```

2. 打断点（真实符号）

```text
break github.com/TripleCGame/go-high-concurrency-learn/internal/service/leak.(*Leaker).leakWorker
continue
```

3. 命中后执行

```text
goroutines
stack
args
```

4. `quit` → `y` 退出。

## 期望现象

- 断点在 `internal/service/leak/leak.go` 的 `leakWorker`
- `goroutines` 里能看到多个 worker
- `stack`：`leakWorker` ← `Start.func1`

## 验收

- [ ] 能独立复现断点命中
- [ ] 能口述：Delve 证明「卡在分配循环」，堆归因还要靠 pprof

## 下一篇

→ [3plan-pprof.md](./3plan-pprof.md)
