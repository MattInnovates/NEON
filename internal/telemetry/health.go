package telemetry

import (
	"sync/atomic"
	"time"
)

// Health holds simple runtime counters
type Health struct {
	startedAt int64
	events    int64
	errors    int64
}

func NewHealth() *Health {
	return &Health{startedAt: time.Now().Unix()}
}

func (h *Health) IncEvents() {
	atomic.AddInt64(&h.events, 1)
}

func (h *Health) IncErrors() {
	atomic.AddInt64(&h.errors, 1)
}

func (h *Health) Snapshot() map[string]any {
	return map[string]any{
		"uptime_sec": time.Now().Unix() - h.startedAt,
		"events":     atomic.LoadInt64(&h.events),
		"errors":     atomic.LoadInt64(&h.errors),
	}
}
