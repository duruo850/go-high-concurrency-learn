package tracex

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// DemoWalk simulates one request's internal call chain for Jaeger waterfall:
// handle → auth → db.query → downstream.rpc
func DemoWalk(ctx context.Context) (traceID string, steps []string) {
	ctx, root := Start(ctx, "trace.demo.handle",
		attribute.String("demo", "trace-walk"),
	)
	defer root.End()

	steps = make([]string, 0, 4)

	{
		c, sp := Start(ctx, "auth.validate")
		time.Sleep(8 * time.Millisecond)
		sp.SetAttributes(attribute.Bool("auth.ok", true))
		sp.End()
		_ = c
		steps = append(steps, "auth.validate")
	}
	{
		c, sp := Start(ctx, "db.query")
		time.Sleep(25 * time.Millisecond)
		sp.SetAttributes(
			attribute.String("db.system", "postgres"),
			attribute.String("db.statement", "SELECT 1 /* demo */"),
		)
		sp.End()
		_ = c
		steps = append(steps, "db.query")
	}
	{
		c, sp := Start(ctx, "downstream.rpc")
		time.Sleep(15 * time.Millisecond)
		sp.SetAttributes(attribute.String("peer.service", "mock-svc"))
		// Demo a recorded error on a child span (does not fail the HTTP request).
		sp.RecordError(context.DeadlineExceeded)
		sp.SetStatus(codes.Error, "demo downstream timeout")
		sp.End()
		_ = c
		steps = append(steps, "downstream.rpc(error)")
	}

	root.SetAttributes(attribute.Int("demo.steps", len(steps)))
	return TraceID(ctx), steps
}
