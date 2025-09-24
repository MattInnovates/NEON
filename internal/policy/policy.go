package policy

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"neon/internal/persona"
)

// Rule defines a simple if-then behavior.
// Conditions: mood + word presence
// Action: override or modify response
type Rule struct {
	When struct {
		Mood string `json:"mood"`
		Word string `json:"word"`
	} `json:"when"`
	Then string `json:"then"` // text template
}

type Engine struct {
	mu    sync.RWMutex
	rules []Rule
	path  string
}

// NewEngine loads rules from a JSON file, or creates empty if not found.
func NewEngine(path string) *Engine {
	e := &Engine{path: path}
	_ = e.Load()
	return e
}

func (e *Engine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	f, err := os.ReadFile(e.path)
	if err != nil {
		e.rules = []Rule{}
		return nil
	}
	var rules []Rule
	if err := json.Unmarshal(f, &rules); err != nil {
		return err
	}
	e.rules = rules
	return nil
}

func (e *Engine) Save() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data, err := json.MarshalIndent(e.rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(e.path, data, 0o644)
}

// Apply checks rules against mood + input text, may return an override response.
func (e *Engine) Apply(mood persona.Mood, input string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, r := range e.rules {
		if (r.When.Mood == "" || strings.EqualFold(r.When.Mood, string(mood))) &&
			(r.When.Word == "" || strings.Contains(strings.ToLower(input), strings.ToLower(r.When.Word))) {
			return r.Then
		}
	}
	return ""
}

// AddRule appends a new rule.
func (e *Engine) AddRule(rule Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// HasRuleFor checks if a rule already exists for a given word.
func (e *Engine) HasRuleFor(word string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, r := range e.rules {
		if strings.EqualFold(r.When.Word, word) {
			return true
		}
	}
	return false
}

// UpdateRule modifies the response text for a given word (if rule exists).
func (e *Engine) UpdateRule(word, newText string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, r := range e.rules {
		if strings.EqualFold(r.When.Word, word) {
			e.rules[i].Then = newText
			return true
		}
	}
	return false
}
