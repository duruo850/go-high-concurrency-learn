// Package main 是学习服务入口：加载配置 → 可观测性初始化 → 泄漏 worker → HTTP → 阻塞运行。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/TripleCGame/go-high-concurrency-learn/internal/conf"
	"github.com/TripleCGame/go-high-concurrency-learn/internal/observability/sentryx"
	"github.com/TripleCGame/go-high-concurrency-learn/internal/observability/tracex"
	"github.com/TripleCGame/go-high-concurrency-learn/internal/pkg/applog"
	httpserver "github.com/TripleCGame/go-high-concurrency-learn/internal/server/http"
	"github.com/TripleCGame/go-high-concurrency-learn/internal/service/echo"
	"github.com/TripleCGame/go-high-concurrency-learn/internal/service/leak"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// flagconf：配置目录，make run 传 -conf configs。
var flagconf string

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf configs")
}

// newApp 组装 Kratos 应用：元数据 + 日志 + HTTP Server。
func newApp(logger log.Logger, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID("go-high-concurrency-learn"),
		kratos.Name("go-high-concurrency-learn"),
		kratos.Version("1.0.0"),
		kratos.Logger(logger),
		kratos.Server(hs),
	)
}

func main() {
	flag.Parse()

	// ---------- 1. 标准输出日志（Kratos 框架用）----------
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service", "go-high-concurrency-learn",
	)

	// ---------- 2. 加载 configs（YAML → conf.Bootstrap）----------
	c := config.New(config.WithSource(file.NewSource(flagconf)))
	defer c.Close()
	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// ---------- 3. 结构化 JSONL 日志（供 Filebeat → ELK）----------
	logPath := "tmp/logs/app.jsonl"
	if bc.Log != nil && bc.Log.Path != "" {
		logPath = bc.Log.Path
	}
	appLog, err := applog.New("go-high-concurrency-learn", logPath)
	if err != nil {
		panic(err)
	}
	defer appLog.Close()
	appLog.Info("server starting", map[string]any{"event": "startup", "conf": flagconf})

	// ---------- 4. Sentry（异常 Issue；DSN 空则跳过）----------
	enabled, err := sentryx.Init(bc.Sentry)
	if err != nil {
		panic(err)
	}
	defer sentryx.Flush()
	if enabled {
		fmt.Println("sentry: enabled")
		appLog.Info("sentry enabled", map[string]any{"event": "sentry_init"})
	} else {
		fmt.Println("sentry: disabled (set sentry.dsn or SENTRY_DSN to enable)")
	}

	// ---------- 5. OpenTelemetry → OTLP（链路追踪；endpoint 空则跳过）----------
	traceOn, err := tracex.Init(bc.Trace)
	if err != nil {
		panic(err)
	}
	defer func() {
		// 进程退出前刷出未发送的 span
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = tracex.Shutdown(ctx)
	}()
	if traceOn {
		fmt.Println("trace: enabled →", tracex.Endpoint())
		appLog.Info("trace enabled", map[string]any{"event": "trace_init", "endpoint": tracex.Endpoint()})
	} else {
		fmt.Println("trace: disabled (set trace.endpoint or OTEL_EXPORTER_OTLP_ENDPOINT)")
	}

	// ---------- 6. 内存泄漏演示 worker（leak / fixed 由配置决定）----------
	leaker := leak.New(bc.Leak)
	leaker.Start()
	defer leaker.Stop()

	// ---------- 7. 业务 + HTTP（health / echo / metrics / pprof / demo APIs）----------
	echoSvc := echo.New()
	hs := httpserver.NewHTTPServer(bc.Server, echoSvc, appLog, logger)
	app := newApp(logger, hs)

	// ---------- 8. 阻塞运行直到信号退出 ----------
	if err := app.Run(); err != nil {
		appLog.Error("server stopped", map[string]any{"event": "shutdown", "error": err.Error()})
		panic(err)
	}
}
