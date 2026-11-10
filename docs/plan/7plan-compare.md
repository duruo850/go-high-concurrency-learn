# 7. 复盘：Prometheus / ELK / Sentry（+ Delve/pprof）

## 目标

形成稳定心智模型，知道线上排障先开哪扇门。

## 对照表

| | Prometheus | ELK | Sentry | Delve / pprof |
|--|------------|-----|--------|----------------|
| 主数据 | 时序指标 | 日志文本 | 错误事件 | 进程调试 / 性能剖析 |
| 采集 | Pull `/metrics` | Push 日志 | SDK 上报错误 | 本机 attach / HTTP pprof |
| 典型问题 | 内存是否持续涨？ | 某请求日志里写了什么？ | 哪个函数 panic？影响面？ | 断点看栈；heap 定因 |
| 本仓库命令 | `make prom-up` | `make elk-up` | `make sentry-boom` | `dlv` / `make pprof-top` |

## 推荐排障顺序（泄漏场景）

1. **Prometheus /metrics**：确认 `go_memstats_*`、`go_goroutines` 异常  
2. **pprof diff**：定位到 `leakWorker`  
3. **Delve**：需要时打断点看现场  
4. **ELK**：若有业务错误日志，检索关联请求  
5. **Sentry**：若是 panic/5xx 异常类问题，看 Issue 聚合  

## 验收（自评）

- [ ] 能不看文档画出三条链路（Prom / ELK / Sentry）
- [ ] 能举一个「只用错工具会怎样」的例子（例如只用 Kibana 查堆泄漏）
- [ ] 本仓库三个实验都至少完整跑过一遍

## 延伸（可选，未强制实现）

- Prometheus → Alertmanager 告警规则（堆阈值）
- Grafana 面板
- Logstash 替代/增强 Filebeat 解析
- 自建 Sentry（Docker）与 Release / Source Maps
