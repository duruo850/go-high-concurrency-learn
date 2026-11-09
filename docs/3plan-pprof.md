# 3. pprof：堆泄漏定因

## 目标

用两次 heap profile 的 diff，把内存增长定位到 `(*Leaker).leakWorker`。

## 前置

- 完成 [2plan-delve.md](./2plan-delve.md)（概念上即可）
- `make run` 在跑，`leak.mode: leak`

## 概念

| 模式 | 用途 |
|------|------|
| 单次 `heap` | 当前 in-use 谁占得多 |
| `-diff_base` 两次对比 | 间隔内**新增**归到谁（本 demo 推荐） |

## 步骤

1. 抓 profile

```bash
make pprof-capture
```

2. 终端定因

```bash
make pprof-top
```

期望 top / list 指向：

```text
.../internal/service/leak.(*Leaker).leakWorker
```

并落到 `buf := make([]byte, l.chunkBytes)`。

3. （可选）UI

```bash
make pprof-diff
```

打开 http://127.0.0.1:8081 。

4. goroutine 文本栈

```bash
curl -sS 'http://127.0.0.1:8000/debug/pprof/goroutine?debug=1' | head -n 30
```

5. 对照：改 `leak.mode: fixed` 重启后再 capture，增长应明显变缓。

## 验收

- [ ] diff top 能指到 `leak.go` 的 `leakWorker`
- [ ] 知道要用 `-diff_base`（Makefile 已封装），不要只靠拍脑袋看单次绝对量

## 下一篇

→ [4plan-prometheus.md](./4plan-prometheus.md)
