package conf

import (
	"fmt"
	"time"
)

type Bootstrap struct {
	Server *Server `json:"server"`
	Leak   *Leak   `json:"leak"`
}

type Server struct {
	HTTP *HTTP `json:"http"`
}

type HTTP struct {
	Network string `json:"network"`
	Addr    string `json:"addr"`
	Timeout string `json:"timeout"`
}

func (h *HTTP) TimeoutDuration() time.Duration {
	if h == nil || h.Timeout == "" {
		return 3 * time.Second
	}
	d, err := time.ParseDuration(h.Timeout)
	if err != nil {
		panic(fmt.Sprintf("invalid server.http.timeout: %v", err))
	}
	return d
}

type Leak struct {
	// Mode is "leak" or "fixed".
	Mode       string `json:"mode"`
	Workers    int    `json:"workers"`
	ChunkBytes int    `json:"chunk_bytes"`
	Interval   string `json:"interval"`
}

func (l *Leak) IntervalDuration() time.Duration {
	if l == nil || l.Interval == "" {
		return 200 * time.Millisecond
	}
	d, err := time.ParseDuration(l.Interval)
	if err != nil {
		panic(fmt.Sprintf("invalid leak.interval: %v", err))
	}
	return d
}
