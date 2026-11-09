package sentryx

import (
	"fmt"
	"os"
	"time"

	"github.com/TripleCGame/go-memleak-demo/internal/conf"
	"github.com/getsentry/sentry-go"
)

// Init configures Sentry from conf and/or SENTRY_DSN env. Empty DSN = no-op enabled=false.
func Init(c *conf.Sentry) (enabled bool, err error) {
	dsn := ""
	env := "local"
	debug := false
	if c != nil {
		dsn = c.DSN
		if c.Environment != "" {
			env = c.Environment
		}
		debug = c.Debug
	}
	if v := os.Getenv("SENTRY_DSN"); v != "" {
		dsn = v
	}
	if dsn == "" {
		return false, nil
	}
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		Debug:            debug,
		TracesSampleRate: 0,
		AttachStacktrace: true,
	})
	if err != nil {
		return false, fmt.Errorf("sentry init: %w", err)
	}
	return true, nil
}

func Flush() {
	sentry.Flush(2 * time.Second)
}

func CaptureException(err error) {
	sentry.CaptureException(err)
}

func CaptureMessage(msg string) {
	sentry.CaptureMessage(msg)
}

func Recover(err any) {
	sentry.CurrentHub().Recover(err)
	Flush()
}
