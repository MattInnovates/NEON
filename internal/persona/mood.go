package persona

import (
	"strings"
	"sync"
	"time"
)

// Mood represents NEON's coarse affective state.
type Mood string

const (
	MoodNeutral  Mood = "neutral"
	MoodPositive Mood = "positive"
	MoodNegative Mood = "negative"
)

// Engine tracks mood over time using a very simple keyword–valence model.
// It’s concurrency-safe and has zero external deps.
type Engine struct {
	mu         sync.RWMutex
	current    Mood
	score      float64 // running sentiment score
	lastUpdate time.Time
	// decay controls how quickly the score drifts back toward zero per second.
	decay float64
}

// NewEngine creates a new mood engine. decay is the score units per second
// pulled toward zero (e.g., 0.05 means ~0.05/sec toward neutral).
func NewEngine(decay float64) *Engine {
	if decay <= 0 {
		decay = 0.05
	}
	return &Engine{
		current:    MoodNeutral,
		score:      0,
		lastUpdate: time.Now(),
		decay:      decay,
	}
}

// Get returns the current mood and a snapshot of the internal score.
func (e *Engine) Get() (Mood, float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.current, e.score
}

// UpdateFromText ingests a user text chunk, applies decay since last update,
// then adjusts the score by keyword valence, and refreshes the discrete mood.
// Returns the new Mood and score.
func (e *Engine) UpdateFromText(text string) (Mood, float64) {
	now := time.Now()

	e.mu.Lock()
	// Apply time decay toward zero
	dt := now.Sub(e.lastUpdate).Seconds()
	if dt > 0 {
		if e.score > 0 {
			e.score = max(0, e.score-e.decay*dt)
		} else if e.score < 0 {
			e.score = min(0, e.score+e.decay*dt)
		}
	}

	// Adjust score based on simple keyword valence.
	delta := sentimentDelta(text)
	e.score += delta

	// Clamp score to a reasonable range so it can’t run away.
	if e.score > 5 {
		e.score = 5
	} else if e.score < -5 {
		e.score = -5
	}

	// Map score to discrete mood
	switch {
	case e.score >= 0.75:
		e.current = MoodPositive
	case e.score <= -0.75:
		e.current = MoodNegative
	default:
		e.current = MoodNeutral
	}

	e.lastUpdate = now
	m, s := e.current, e.score
	e.mu.Unlock()
	return m, s
}

// sentimentDelta computes a tiny valence delta from the given text.
func sentimentDelta(text string) float64 {
	t := strings.ToLower(text)

	positives := [...]string{
		"good", "great", "awesome", "love", "nice", "cool", "amazing", "yay",
		"thanks", "thank you", "excellent", "perfect", "happy", "success",
	}
	negatives := [...]string{
		"bad", "terrible", "awful", "hate", "annoying", "broken", "sad",
		"angry", "fail", "failure", "bug", "crash", "worse", "worst",
	}

	score := 0.0
	for _, p := range positives {
		if strings.Contains(t, p) {
			score += 0.5
		}
	}
	for _, n := range negatives {
		if strings.Contains(t, n) {
			score -= 0.5
		}
	}

	// Amplifier for !!!
	if strings.Contains(t, "!!!") {
		if score > 0 {
			score += 0.25
		} else if score < 0 {
			score -= 0.25
		}
	}
	// Soften negative if phrased as a question
	if strings.HasSuffix(t, "?") && score < 0 {
		score *= 0.8
	}
	return score
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
