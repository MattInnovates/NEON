package storage

import (
	"sort"
	"strings"
	"sync"
)

// Weights holds word → count mappings with safe concurrency.
type Weights struct {
	mu    sync.RWMutex
	count map[string]int
	dirty bool
}

// NewWeights returns an empty weights map.
func NewWeights() *Weights {
	return &Weights{
		count: make(map[string]int),
	}
}

// Update increments counts for all words in the text.
func (w *Weights) Update(text string) {
	words := Tokenize(text) // <-- capital T now
	w.mu.Lock()
	for _, tok := range words {
		w.count[tok]++
	}
	w.dirty = true
	w.mu.Unlock()
}

// TopN returns the top-N words sorted by frequency.
func (w *Weights) TopN(n int) []WordCount {
	w.mu.RLock()
	defer w.mu.RUnlock()

	pairs := make([]WordCount, 0, len(w.count))
	for k, v := range w.count {
		pairs = append(pairs, WordCount{Word: k, Count: v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	if n > 0 && n < len(pairs) {
		return pairs[:n]
	}
	return pairs
}

// Snapshot returns a copy of the current word counts.
func (w *Weights) Snapshot() map[string]int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	copyMap := make(map[string]int, len(w.count))
	for k, v := range w.count {
		copyMap[k] = v
	}
	return copyMap
}

// Save persists the weights using AtomicWriteJSON if dirty.
func (w *Weights) Save(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.dirty {
		return nil
	}
	if err := AtomicWriteJSON(path, w.count); err != nil {
		return err
	}
	w.dirty = false
	return nil
}

// Load replaces current weights with those from disk.
func (w *Weights) Load(path string) error {
	m := make(map[string]int)
	if !Exists(path) {
		w.mu.Lock()
		w.count = m
		w.mu.Unlock()
		return nil
	}
	if err := ReadJSON(path, &m); err != nil {
		return err
	}
	w.mu.Lock()
	w.count = m
	w.mu.Unlock()
	return nil
}

// WordCount is a helper struct for TopN output.
type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// Tokenize splits text into lowercase tokens.
func Tokenize(s string) []string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer(
		".", " ", ",", " ", "!", " ", "?", " ",
		"(", " ", ")", " ", "[", " ", "]", " ",
		"\"", " ", "'", " ", ":", " ", ";", " ",
	)
	s = repl.Replace(s)
	return strings.Fields(s)
}
