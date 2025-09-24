package storage

import (
	"fmt"
	"path/filepath"
	"time"
)

// Snapshot captures NEON's state for rollback.
type Snapshot struct {
	Schema    int                `json:"schema"`
	Timestamp time.Time          `json:"timestamp"`
	Self      map[string]any     `json:"self"`
	Beliefs   []string           `json:"beliefs"`
	Weights   map[string]float64 `json:"weights"`
	Features  map[string]bool    `json:"features"`
	Notes     string             `json:"notes"`
}

const snapshotSchema = 1

func snapshotPath(root string, t time.Time) string {
	name := "run-" + t.UTC().Format("2006-01-02T15-04-05Z") + ".json"
	return filepath.Join(root, "data", "snapshots", name)
}

func SaveSnapshot(root string, self map[string]any, beliefs []string, weights map[string]float64, features map[string]bool, notes string) (string, error) {
	s := Snapshot{
		Schema:    snapshotSchema,
		Timestamp: time.Now().UTC(),
		Self:      self,
		Beliefs:   beliefs,
		Weights:   weights,
		Features:  features,
		Notes:     notes,
	}
	p := snapshotPath(root, s.Timestamp)
	return p, AtomicWriteJSON(p, &s)
}

func LoadSnapshot(path string) (*Snapshot, error) {
	var s Snapshot
	if err := ReadJSON(path, &s); err != nil {
		return nil, err
	}
	if s.Schema != snapshotSchema {
		return nil, fmt.Errorf("snapshot schema mismatch: got %d", s.Schema)
	}
	return &s, nil
}
