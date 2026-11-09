package leak

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/TripleCGame/go-memleak-demo/internal/conf"
)

// Leaker spawns background workers that either leak or stay bounded.
type Leaker struct {
	mode       string
	workers    int
	chunkBytes int
	interval   time.Duration

	mu     sync.Mutex
	held   [][]byte // retained only in leak mode
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(c *conf.Leak) *Leaker {
	mode := "leak"
	workers := 8
	chunk := 64 * 1024
	interval := 200 * time.Millisecond
	if c != nil {
		if c.Mode != "" {
			mode = c.Mode
		}
		if c.Workers > 0 {
			workers = c.Workers
		}
		if c.ChunkBytes > 0 {
			chunk = c.ChunkBytes
		}
		if c.Interval != "" {
			interval = c.IntervalDuration()
		}
	}
	return &Leaker{
		mode:       mode,
		workers:    workers,
		chunkBytes: chunk,
		interval:   interval,
	}
}

// Start launches workers immediately. Safe to call once at process start.
func (l *Leaker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel
	log.Printf("leak: starting mode=%s workers=%d chunk=%d interval=%s",
		l.mode, l.workers, l.chunkBytes, l.interval)
	for i := 0; i < l.workers; i++ {
		id := i
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			if l.mode == "fixed" {
				l.fixedWorker(ctx, id)
				return
			}
			l.leakWorker(ctx, id)
		}()
	}
}

// Stop cancels workers and waits for them to exit (mainly useful in fixed mode / tests).
func (l *Leaker) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
	l.wg.Wait()
}

// HeldBytes reports retained heap from leak mode (for tests / delve inspection).
func (l *Leaker) HeldBytes() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for _, b := range l.held {
		n += len(b)
	}
	return n
}

// leakWorker never releases allocations and never exits until process stop.
// This is the intentional goroutine + heap leak for the demo.
func (l *Leaker) leakWorker(ctx context.Context, id int) {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			buf := make([]byte, l.chunkBytes)
			for i := range buf {
				buf[i] = byte(id)
			}
			l.mu.Lock()
			l.held = append(l.held, buf)
			l.mu.Unlock()
		}
	}
}

// fixedWorker allocates temporarily then drops references so GC can reclaim.
func (l *Leaker) fixedWorker(ctx context.Context, id int) {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			buf := make([]byte, l.chunkBytes)
			for i := range buf {
				buf[i] = byte(id)
			}
			_ = buf // intentionally drop; do not retain
		}
	}
}
