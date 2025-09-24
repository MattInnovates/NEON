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
}

func NewAgent(logger *telemetry.Logger) *Agent {
	weights := storage.NewWeights()
	path := filepath.Join("data", "beliefs", "weights.json")
	_ = os.MkdirAll(filepath.Dir(path), 0o755)

	if err := weights.Load(path); err != nil {
		fmt.Println("⚠ failed to load weights:", err)
	}

	return &Agent{
		logger:    logger,
		mood:      persona.NewEngine(0.05),
		weights:   weights,
		path:      path,
		cognition: cognition.NewEngine(weights),
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
			if line == "exit" {
				a.logger.Log(structs.NewEvent("EXIT", "console", map[string]any{
					"message": "User requested shutdown",
				}))

				if err := a.weights.Save(a.path); err != nil {
					fmt.Println("⚠ failed to save weights:", err)
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

			// Log INPUT
			a.logger.Log(structs.NewEvent("INPUT", "console", map[string]any{
				"text":       line,
				"mood_now":   string(curMood),
				"mood_score": curScore,
				"top_words":  a.weights.TopN(5),
			}))

			// Generate normal response
			resp := a.cognition.Respond(line, curMood)
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
