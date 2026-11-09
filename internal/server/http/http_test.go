package http_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TripleCGame/go-memleak-demo/internal/conf"
	"github.com/TripleCGame/go-memleak-demo/internal/pkg/applog"
	httpserver "github.com/TripleCGame/go-memleak-demo/internal/server/http"
	"github.com/TripleCGame/go-memleak-demo/internal/service/echo"
	"github.com/go-kratos/kratos/v2/log"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	appLog, err := applog.New("test", filepath.Join(dir, "app.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = appLog.Close() })

	srv := httpserver.NewHTTPServer(
		&conf.Server{HTTP: &conf.HTTP{Addr: "127.0.0.1:0", Timeout: "2s"}},
		echo.New(),
		appLog,
		log.NewStdLogger(io.Discard),
	)
	return httptest.NewServer(srv)
}

func TestHealth(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body = %#v", body)
	}
}

func TestEchoAPI(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/echo?msg=hello")
	if err != nil {
		t.Fatalf("GET /api/echo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["message"] != "hello" {
		t.Fatalf("body = %#v, want message=hello", body)
	}

	resp2, err := http.Get(ts.URL + "/api/echo")
	if err != nil {
		t.Fatalf("GET /api/echo default: %v", err)
	}
	defer resp2.Body.Close()
	var body2 map[string]string
	if err := json.NewDecoder(resp2.Body).Decode(&body2); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body2["message"] != "pong" {
		t.Fatalf("default body = %#v, want message=pong", body2)
	}
}

func TestLogDemoWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.jsonl")
	appLog, err := applog.New("test", path)
	if err != nil {
		t.Fatal(err)
	}
	defer appLog.Close()

	srv := httpserver.NewHTTPServer(
		&conf.Server{HTTP: &conf.HTTP{Addr: "127.0.0.1:0", Timeout: "2s"}},
		echo.New(),
		appLog,
		log.NewStdLogger(io.Discard),
	)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/log-demo?topic=unit")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"event":"log_demo"`) || !strings.Contains(s, `"topic":"unit"`) {
		t.Fatalf("log file missing fields:\n%s", s)
	}
}

func TestMetricsAndPprof(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics status = %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	s := string(b)
	if !strings.Contains(s, "go_goroutines") || !strings.Contains(s, "go_memstats_heap_alloc_bytes") {
		snippet := s
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		t.Fatalf("metrics missing expected series:\n%s", snippet)
	}

	resp2, err := http.Get(ts.URL + "/debug/pprof/heap")
	if err != nil {
		t.Fatalf("GET /debug/pprof/heap: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("pprof heap status = %d", resp2.StatusCode)
	}
}
