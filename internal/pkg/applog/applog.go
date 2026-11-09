package applog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger writes one JSON object per line (stdout + optional file for Filebeat).
type Logger struct {
	service string
	mu      sync.Mutex
	out     io.Writer
	file    *os.File
}

func New(service, path string) (*Logger, error) {
	l := &Logger{service: service, out: os.Stdout}
	if path == "" {
		return l, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir log dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	l.file = f
	l.out = io.MultiWriter(os.Stdout, f)
	return l, nil
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) Info(msg string, fields map[string]any)  { l.log("info", msg, fields) }
func (l *Logger) Warn(msg string, fields map[string]any)  { l.log("warn", msg, fields) }
func (l *Logger) Error(msg string, fields map[string]any) { l.log("error", msg, fields) }

func (l *Logger) log(level, msg string, fields map[string]any) {
	if l == nil {
		return
	}
	rec := map[string]any{
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"level":   level,
		"msg":     msg,
		"service": l.service,
	}
	for k, v := range fields {
		rec[k] = v
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.out.Write(append(b, '\n'))
}
