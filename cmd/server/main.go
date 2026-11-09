package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TripleCGame/go-memleak-demo/internal/conf"
	"github.com/TripleCGame/go-memleak-demo/internal/observability/sentryx"
	"github.com/TripleCGame/go-memleak-demo/internal/pkg/applog"
	httpserver "github.com/TripleCGame/go-memleak-demo/internal/server/http"
	"github.com/TripleCGame/go-memleak-demo/internal/service/echo"
	"github.com/TripleCGame/go-memleak-demo/internal/service/leak"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

var flagconf string

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf configs")
}

func newApp(logger log.Logger, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID("go-memleak-demo"),
		kratos.Name("go-memleak-demo"),
		kratos.Version("1.0.0"),
		kratos.Logger(logger),
		kratos.Server(hs),
	)
}

func main() {
	flag.Parse()

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service", "go-memleak-demo",
	)

	c := config.New(config.WithSource(file.NewSource(flagconf)))
	defer c.Close()
	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logPath := "tmp/logs/app.jsonl"
	if bc.Log != nil && bc.Log.Path != "" {
		logPath = bc.Log.Path
	}
	appLog, err := applog.New("go-memleak-demo", logPath)
	if err != nil {
		panic(err)
	}
	defer appLog.Close()
	appLog.Info("server starting", map[string]any{"event": "startup", "conf": flagconf})

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

	leaker := leak.New(bc.Leak)
	leaker.Start()
	defer leaker.Stop()

	echoSvc := echo.New()
	hs := httpserver.NewHTTPServer(bc.Server, echoSvc, appLog, logger)
	app := newApp(logger, hs)

	if err := app.Run(); err != nil {
		appLog.Error("server stopped", map[string]any{"event": "shutdown", "error": err.Error()})
		panic(err)
	}
}
