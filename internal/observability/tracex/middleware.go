package tracex

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPFilter wraps the whole HTTP stack with otelhttp (server spans + W3C propagation).
// Safe when tracing is disabled (noop provider).
func HTTPFilter(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http.server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}
