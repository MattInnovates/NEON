package cognition

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"neon/internal/persona"
	"neon/internal/storage"
)

type Engine struct {
	weights *storage.Weights
	rng     *rand.Rand
	seen    map[string]bool // track seen words for novelty
}

func NewEngine(weights *storage.Weights) *Engine {
	return &Engine{
		weights: weights,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		seen:    make(map[string]bool),
	}
}

func (e *Engine) Respond(userText string, mood persona.Mood) string {
	top := e.weights.TopN(5)
	words := make([]string, 0, len(top))
	for _, wc := range top {
		words = append(words, wc.Word)
	}

	var base string
	switch mood {
	case persona.MoodPositive:
		base = "I like"
	case persona.MoodNegative:
		base = "I don't like"
	default:
		base = "I know"
	}

	if len(words) > 0 {
		choice := words[e.rng.Intn(len(words))]
		return fmt.Sprintf("(%s) %s %s", mood, base, choice)
	}
	return fmt.Sprintf("(%s) You said: %s", mood, userText)
}

// ReflectIfNeeded may return a reflection, or "" if no reflection occurs.
func (e *Engine) ReflectIfNeeded(userText string, mood persona.Mood, score float64) string {
	words := storage.Tokenize(userText)
	novelty := false
	for _, w := range words {
		if !e.seen[w] {
			e.seen[w] = true
			novelty = true
		}
	}

	// Conditions for reflection
	trigger := false

	// 1. New word discovered
	if novelty {
		trigger = true
	}

	// 2. Mood extremes
	if score >= 2.0 || score <= -2.0 {
		if e.rng.Float64() < 0.5 { // 50% chance
			trigger = true
		}
	}

	// 3. Small random chance always
	if e.rng.Intn(6) == 0 { // ~1 in 6
		trigger = true
	}

	if !trigger {
		return ""
	}

	// Build reflection from top memory
	top := e.weights.TopN(3)
	if len(top) == 0 {
		return fmt.Sprintf("(%s) I don't know much yet.", mood)
	}

	wordsOut := make([]string, 0, len(top))
	for _, wc := range top {
		wordsOut = append(wordsOut, wc.Word)
	}
	return fmt.Sprintf("(%s) I keep thinking about: %s", mood, strings.Join(wordsOut, ", "))
}
