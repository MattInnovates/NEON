package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"neon/pkg/structs"
)

// Logger centralizes event logging with daily rotation
type Logger struct {
	mu     sync.Mutex
	file   *os.File
	enc    *json.Encoder
	day    string
	root   string
	events chan structs.Event
	stop   chan struct{}
	health *Health
}

// NewLogger creates and starts a logger goroutine
func NewLogger(root string) *Logger {
	l := &Logger{
		root:   root,
		events: make(chan structs.Event, 100),
		stop:   make(chan struct{}),
		health: NewHealth(),
	}
	go l.loop()
	go l.periodicHealth() // emit HEALTH snapshots every 30s
	return l
}

func (l *Logger) loop() {
	for {
		select {
		case ev := <-l.events:
			_ = l.write(ev)
		case <-l.stop:
			if l.file != nil {
				_ = l.file.Close()
			}
			return
		}
	}
}

func (l *Logger) write(ev structs.Event) error {
	day := time.Now().Format("2006-01-02")
	if l.day != day {
		if l.file != nil {
			_ = l.file.Close()
		}
		path := filepath.Join(l.root, "data", "events", fmt.Sprintf("events-%s.jsonl", day))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		l.file = f
		l.enc = json.NewEncoder(f)
		l.day = day
	}
	return l.enc.Encode(ev)
}

// Log queues an event for writing
func (l *Logger) Log(ev structs.Event) {
	l.health.IncEvents()
	select {
	case l.events <- ev:
	default:
		// channel full, drop event
		l.health.IncErrors()
	}
}

// Health returns a snapshot of runtime counters
func (l *Logger) Health() map[string]any {
	return l.health.Snapshot()
}

// periodicHealth emits a HEALTH event every 30s
func (l *Logger) periodicHealth() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-ticker.C:
			ev := structs.NewEvent("HEALTH", "system", l.Health())
			l.Log(ev)
		}
	}
}

// Close stops the logger goroutine
func (l *Logger) Close() {
	close(l.stop)
}
