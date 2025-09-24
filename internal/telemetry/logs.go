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

type Logger struct {
	mu     sync.Mutex
	file   *os.File
	enc    *json.Encoder
	day    string
	root   string
	events chan structs.Event
	stop   chan struct{}
}

func NewLogger(root string) *Logger {
	l := &Logger{
		root:   root,
		events: make(chan structs.Event, 100),
		stop:   make(chan struct{}),
	}
	go l.loop()
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

func (l *Logger) Log(ev structs.Event) {
	select {
	case l.events <- ev:
	default:
		// drop if channel full
	}
}

func (l *Logger) Close() {
	close(l.stop)
}
