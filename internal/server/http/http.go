package http

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/TripleCGame/go-memleak-demo/internal/conf"
	"github.com/TripleCGame/go-memleak-demo/internal/observability/sentryx"
	"github.com/TripleCGame/go-memleak-demo/internal/pkg/applog"
	"github.com/TripleCGame/go-memleak-demo/internal/service/echo"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPServer wires health, metrics, pprof, echo, ELK demo log and Sentry demo APIs.
func NewHTTPServer(c *conf.Server, echoSvc *echo.Service, appLog *applog.Logger, logger log.Logger) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(
			recovery.Recovery(recovery.WithHandler(func(ctx context.Context, _, err any) error {
				sentryx.Recover(err)
				if appLog != nil {
					appLog.Error("panic recovered", map[string]any{
						"error": fmt.Sprint(err),
					})
				}
				return errors.InternalServer("PANIC", fmt.Sprint(err))
			})),
		),
		khttp.Logger(logger),
	}
	if c != nil && c.HTTP != nil {
		if c.HTTP.Network != "" {
			opts = append(opts, khttp.Network(c.HTTP.Network))
		}
		if c.HTTP.Addr != "" {
			opts = append(opts, khttp.Address(c.HTTP.Addr))
		}
		if d := c.HTTP.TimeoutDuration(); d > 0 {
			opts = append(opts, khttp.Timeout(d))
		}
	}

	srv := khttp.NewServer(opts...)
	r := srv.Route("/")

	r.GET("/health", func(ctx khttp.Context) error {
		return ctx.Result(http.StatusOK, map[string]string{"status": "ok"})
	})

	r.GET("/api/echo", func(ctx khttp.Context) error {
		msg := ctx.Query().Get("msg")
		out := echoSvc.Echo(ctx, msg)
		if appLog != nil {
			appLog.Info("echo", map[string]any{"msg": out, "event": "echo"})
		}
		return ctx.Result(http.StatusOK, map[string]string{"message": out})
	})

	// ELK learning: emit a structured log line intentionally.
	r.GET("/api/log-demo", func(ctx khttp.Context) error {
		topic := ctx.Query().Get("topic")
		if topic == "" {
			topic = "log_demo"
		}
		if appLog != nil {
			appLog.Info("elk sample event", map[string]any{
				"event": "log_demo",
				"topic": topic,
			})
		}
		return ctx.Result(http.StatusOK, map[string]string{
			"ok":    "true",
			"topic": topic,
			"hint":  "check tmp/logs/app.jsonl and Kibana Discover",
		})
	})

	// Sentry learning: capture exception without crashing process.
	r.GET("/api/boom", func(ctx khttp.Context) error {
		err := errors.New(500, "DEMO_BOOM", "intentional sentry demo error")
		sentryx.CaptureException(err)
		if appLog != nil {
			appLog.Error("sentry boom", map[string]any{"event": "sentry_boom", "error": err.Error()})
		}
		sentryx.Flush()
		return err
	})

	// Sentry learning: panic → local recover → Sentry（Route 不会自动套 Middleware Recovery）.
	r.GET("/api/boom-panic", func(ctx khttp.Context) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				sentryx.Recover(rerr)
				if appLog != nil {
					appLog.Error("panic recovered", map[string]any{
						"event": "sentry_boom_panic",
						"error": fmt.Sprint(rerr),
					})
				}
				err = errors.InternalServer("PANIC", fmt.Sprint(rerr))
			}
		}()
		panic("intentional panic for sentry recovery demo")
	})

	r.GET("/api/sentry-message", func(ctx khttp.Context) error {
		msg := ctx.Query().Get("msg")
		if msg == "" {
			msg = "hello-sentry"
		}
		sentryx.CaptureMessage(msg)
		if appLog != nil {
			appLog.Info("sentry message", map[string]any{"event": "sentry_message", "msg": msg})
		}
		sentryx.Flush()
		return ctx.Result(http.StatusOK, map[string]string{"sent": msg})
	})

	srv.Handle("/metrics", promhttp.Handler())

	srv.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/trace", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/goroutine", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/allocs", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/block", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/mutex", http.DefaultServeMux.ServeHTTP)
	srv.HandleFunc("/debug/pprof/threadcreate", http.DefaultServeMux.ServeHTTP)

	return srv
}
