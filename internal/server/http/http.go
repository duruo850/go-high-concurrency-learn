package http

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/TripleCGame/go-memleak-demo/internal/conf"
	"github.com/TripleCGame/go-memleak-demo/internal/service/echo"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPServer wires health, metrics, pprof and the echo API.
func NewHTTPServer(c *conf.Server, echoSvc *echo.Service, logger log.Logger) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(recovery.Recovery()),
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
		return ctx.Result(http.StatusOK, map[string]string{
			"message": echoSvc.Echo(ctx, msg),
		})
	})

	// Prometheus metrics (includes go_goroutines, go_memstats_*)
	srv.Handle("/metrics", promhttp.Handler())

	// pprof endpoints under /debug/pprof/*
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
