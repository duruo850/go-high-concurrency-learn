APP      := go-high-concurrency-learn
ADDR     := 127.0.0.1:8000
BASE     := http://$(ADDR)
CONF     := configs
PPROFDIR := tmp/pprof
LOGDIR   := tmp/logs

.PHONY: run test echo metrics-watch \
	prom-up prom-down elk-up elk-down elk-sample obs-up obs-down \
	pprof-capture pprof-diff pprof-top pprof-heap \
	sentry-boom sentry-message tidy

tidy:
	go mod tidy

run:
	@mkdir -p $(LOGDIR)
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
	docker compose --profile prom up -d
	@echo "Prometheus UI: http://127.0.0.1:9090"
	@echo "Targets:      http://127.0.0.1:9090/targets"

prom-down:
	docker compose --profile prom down

elk-up:
	@mkdir -p $(LOGDIR)
	docker compose --profile elk up -d
	@echo "Elasticsearch: http://127.0.0.1:9200"
	@echo "Kibana:        http://127.0.0.1:5601"
	@echo "Filebeat reads $(LOGDIR)/*.jsonl → index go-high-concurrency-learn-*"

elk-down:
	docker compose --profile elk down

elk-sample:
	@mkdir -p $(LOGDIR)
	@echo "==> seed logs via /api/echo and /api/log-demo"
	@curl -sS "$(BASE)/api/echo?msg=elk-hello" ; echo
	@curl -sS "$(BASE)/api/log-demo?topic=log_demo" ; echo
	@echo "==> last log lines:"
	@tail -n 5 $(LOGDIR)/app.jsonl 2>/dev/null || echo "(no log file yet — is make run up?)"

obs-up:
	@mkdir -p $(LOGDIR)
	docker compose --profile all up -d
	@echo "Prometheus :9090 | Kibana :5601 | ES :9200"

obs-down:
	docker compose --profile all down

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

sentry-boom:
	@echo "==> GET /api/boom (CaptureException)"
	@curl -sS -w "\nHTTP %{http_code}\n" $(BASE)/api/boom || true
	@echo "==> GET /api/boom-panic (panic → recovery → Sentry)"
	@curl -sS -w "\nHTTP %{http_code}\n" $(BASE)/api/boom-panic || true

sentry-message:
	@curl -sS -w "\nHTTP %{http_code}\n" "$(BASE)/api/sentry-message?msg=hello-sentry"
