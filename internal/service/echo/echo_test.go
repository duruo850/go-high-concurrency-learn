package echo

import (
	"context"
	"testing"
)

func TestEcho(t *testing.T) {
	s := New()
	if got := s.Echo(context.Background(), ""); got != "pong" {
		t.Fatalf("default = %q, want pong", got)
	}
	if got := s.Echo(context.Background(), "hello"); got != "hello" {
		t.Fatalf("msg = %q, want hello", got)
	}
}
