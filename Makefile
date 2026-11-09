APP      := go-memleak-demo
ADDR     := 127.0.0.1:8000
BASE     := http://$(ADDR)
CONF     := configs
PPROFDIR := tmp/pprof

.PHONY: run test echo metrics-watch prom-up prom-down pprof-capture pprof-diff pprof-top pprof-heap tidy

tidy:
	go mod tidy

run:
	go run ./cmd/server -conf $(CONF)

test:
	go test ./...

echo:
	@echo "==> GET /health"
	@curl -sS -w "\nHTTP %{http_code}\n" $(BASE)/health
	@echo "==> GET /api/echo?msg=hello"
	@curl -sS -w "\nHTTP %{http_code}\n" "$(BASE)/api/echo?msg=hello"
	@echo "==> GET /api/echo (default pong)"
	@curl -sS -w "\nHTTP %{http_code}\n" $(BASE)/api/echo

metrics-watch:
	@echo "Polling go_goroutines / go_memstats_heap_alloc_bytes every 5s (Ctrl-C to stop)"
	@while true; do \
		date '+%H:%M:%S'; \
		curl -sS $(BASE)/metrics | grep -E '^go_goroutines |^go_memstats_heap_alloc_bytes '; \
		echo; \
		sleep 5; \
	done

prom-up:
	docker compose up -d
	@echo "Prometheus UI: http://127.0.0.1:9090"
	@echo "Targets:      http://127.0.0.1:9090/targets"

prom-down:
	docker compose down

pprof-capture:
	@mkdir -p $(PPROFDIR)
	@echo "Capture h1.out ..."
	@curl -sS $(BASE)/debug/pprof/heap -o $(PPROFDIR)/h1.out
	@echo "Sleep 20s while leak grows ..."
	@sleep 20
	@echo "Capture h2.out ..."
	@curl -sS $(BASE)/debug/pprof/heap -o $(PPROFDIR)/h2.out
	@ls -la $(PPROFDIR)/h1.out $(PPROFDIR)/h2.out

pprof-diff:
	@test -f $(PPROFDIR)/h1.out && test -f $(PPROFDIR)/h2.out || (echo "run make pprof-capture first"; exit 1)
	@echo "Opening pprof UI on http://127.0.0.1:8081 (use -diff_base so growth attributes to leakWorker)"
	go tool pprof -http=127.0.0.1:8081 -diff_base=$(PPROFDIR)/h1.out $(PPROFDIR)/h2.out

pprof-top:
	@test -f $(PPROFDIR)/h1.out && test -f $(PPROFDIR)/h2.out || (echo "run make pprof-capture first"; exit 1)
	go tool pprof -top -diff_base=$(PPROFDIR)/h1.out $(PPROFDIR)/h2.out
	@echo "---- list leakWorker ----"
	go tool pprof -list='.*leakWorker' -diff_base=$(PPROFDIR)/h1.out $(PPROFDIR)/h2.out

pprof-heap: pprof-capture pprof-diff
