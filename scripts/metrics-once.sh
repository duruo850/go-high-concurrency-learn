#!/usr/bin/env bash
set -euo pipefail
BASE="${1:-http://127.0.0.1:8000}"
echo "== sample 1 =="
curl -sS "$BASE/metrics" | grep -E '^go_goroutines |^go_memstats_heap_alloc_bytes '
echo
echo "sleep 30s ..."
sleep 30
echo "== sample 2 (expect larger heap / high goroutines) =="
curl -sS "$BASE/metrics" | grep -E '^go_goroutines |^go_memstats_heap_alloc_bytes '
