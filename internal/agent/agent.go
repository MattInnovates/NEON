package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"neon/internal/cognition"
	"neon/internal/persona"
	"neon/internal/policy"
	"neon/internal/storage"
	"neon/internal/telemetry"
	"neon/pkg/structs"
)

type Agent struct {
	logger    *telemetry.Logger
	mood      *persona.Engine
	weights   *storage.Weights
	path      string
	cognition *cognition.Engine
	policy    *policy.Engine
}

func NewAgent(logger *telemetry.Logger) *Agent {
	// Word weights (beliefs)
	weights := storage.NewWeights()
	weightsPath := filepath.Join("data", "beliefs", "weights.json")
	_ = os.MkdirAll(filepath.Dir(weightsPath), 0o755)
	_ = weights.Load(weightsPath)

	// Policy rules
	policyPath := filepath.Join("data", "policy", "policy.json")
	_ = os.MkdirAll(filepath.Dir(policyPath), 0o755)
	policies := policy.NewEngine(policyPath)

	return &Agent{
		logger:    logger,
		mood:      persona.NewEngine(0.05),
		weights:   weights,
		path:      weightsPath,
		cognition: cognition.NewEngine(weights),
		policy:    policies,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	a.logger.Log(structs.NewEvent("BOOT", "system", map[string]any{
		"message": "NEON boot sequence",
	}))

	fmt.Println("Type something (or 'exit' to quit):")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Print(">> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)

			// Exit condition
			if line == "exit" {
				a.logger.Log(structs.NewEvent("EXIT", "console", map[string]any{
					"message": "User requested shutdown",
				}))

				// Save beliefs
				if err := a.weights.Save(a.path); err != nil {
					fmt.Println("⚠ failed to save weights:", err)
				}
				// Save policy rules
				if err := a.policy.Save(); err != nil {
					fmt.Println("⚠ failed to save policy:", err)
				}

				fmt.Println("Goodbye.")
				health := a.logger.Health()
				fmt.Printf("Health summary: uptime=%ds, events=%d, errors=%d\n",
					health["uptime_sec"], health["events"], health["errors"])
				return nil
			}

			// Update mood
			curMood, curScore := a.mood.UpdateFromText(line)

			// Update weights
			a.weights.Update(line)
			_ = a.weights.Save(a.path)

			// Check for new/high-frequency words and propose/edit rules
			for _, wc := range a.weights.TopN(3) {
				if !a.policy.HasRuleFor(wc.Word) {
					// New word → propose new rule
					r := policy.Rule{}
					r.When.Mood = string(curMood)
					r.When.Word = wc.Word
					r.Then = fmt.Sprintf("I noticed the word '%s'.", wc.Word)

					a.policy.AddRule(r)
					_ = a.policy.Save()

					a.logger.Log(structs.NewEvent("PROPOSE", "agent", map[string]any{
						"word":  wc.Word,
						"mood":  curMood,
						"rule":  r,
						"count": wc.Count,
					}))

					fmt.Printf("(%s) I created a new rule for '%s'.\n", curMood, wc.Word)

				} else {
					// Existing rule → maybe edit if mood context has shifted
					newText := fmt.Sprintf("Now I feel %s about '%s'.", curMood, wc.Word)
					if a.policy.UpdateRule(wc.Word, newText) {
						_ = a.policy.Save()
						a.logger.Log(structs.NewEvent("EDIT", "agent", map[string]any{
							"word":  wc.Word,
							"mood":  curMood,
							"text":  newText,
							"count": wc.Count,
						}))
						fmt.Printf("(%s) I updated my rule for '%s'.\n", curMood, wc.Word)
					}
				}
			}

			// Log INPUT
			a.logger.Log(structs.NewEvent("INPUT", "console", map[string]any{
				"text":       line,
				"mood_now":   string(curMood),
				"mood_score": curScore,
				"top_words":  a.weights.TopN(5),
			}))

			// Generate cognition-based response
			resp := a.cognition.Respond(line, curMood)

			// Apply policy override if matched
			if override := a.policy.Apply(curMood, line); override != "" {
				resp = fmt.Sprintf("(%s) %s", curMood, override)
			}

			// Output & log
			fmt.Println(resp)
			a.logger.Log(structs.NewEvent("OUTPUT", "agent", map[string]any{
				"text":       resp,
				"mood_now":   string(curMood),
				"mood_score": curScore,
			}))

			// Maybe reflect
			if refl := a.cognition.ReflectIfNeeded(line, curMood, curScore); refl != "" {
				fmt.Println(refl)
				a.logger.Log(structs.NewEvent("REFLECT", "agent", map[string]any{
					"text":       refl,
					"mood_now":   string(curMood),
					"mood_score": curScore,
				}))
			}

			time.Sleep(30 * time.Millisecond)
		}
	}
}
