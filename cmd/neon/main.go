package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "time"
)

// Placeholder Proposal type (real one in internal/selfmod)
type Proposal struct {
    Schema int               `json:"schema"`
    Notes  string            `json:"notes"`
    Flags  map[string]bool   `json:"flags,omitempty"`
}

func main() {
    sandbox := flag.Bool("sandbox", false, "run in sandbox mode (read Proposal from stdin, dry-run only)")
    replay := flag.String("replay", "", "replay events file and exit")
    flag.Parse()

    if *sandbox {
        runSandboxMode()
        return
    }
    if *replay != "" {
        fmt.Println("Replay mode not yet implemented")
        return
    }

    ctx := context.Background()
    agent := NewAgent()
    if err := agent.Run(ctx); err != nil {
        log.Fatalf("agent error: %v", err)
    }
}

// stub
type Agent struct{}
func NewAgent() *Agent { return &Agent{} }
func (a *Agent) Run(ctx context.Context) error {
    fmt.Println("⟦ NEON booting ⟧")
    time.Sleep(1 * time.Second)
    fmt.Println("NEON loop would run here...")
    return nil
}

func runSandboxMode() {
    dec := json.NewDecoder(os.Stdin)
    var p Proposal
    if err := dec.Decode(&p); err != nil {
        json.NewEncoder(os.Stdout).Encode(map[string]any{"accepted": false, "reason": err.Error()})
        os.Exit(1)
    }
    json.NewEncoder(os.Stdout).Encode(map[string]any{"accepted": true})
}
