package echo

import (
	"context"
)

// Service handles the demo business endpoint.
type Service struct{}

func New() *Service {
	return &Service{}
}

// Echo returns the given message, defaulting to "pong".
func (s *Service) Echo(_ context.Context, msg string) string {
	if msg == "" {
		return "pong"
	}
	return msg
}
