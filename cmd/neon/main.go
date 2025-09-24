package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"neon/internal/storage"
	"neon/internal/telemetry"
	"neon/pkg/structs"
)

type Agent struct {
	logger *telemetry.Logger
}

func NewAgent() *Agent {
	return &Agent{logger: telemetry.NewLogger(".")}
}

func (a *Agent) Run(ctx context.Context) error {
	defer a.logger.Close()

	// Boot log
	a.logger.Log(structs.NewEvent("BOOT", "system", map[string]any{
		"message": "NEON boot sequence",
	}))

	fmt.Println("⟦ NEON v0.1 ⟧")
	fmt.Println("NEON is running. Logs are being written to data/events/")
	return nil
}

func main() {
	cmd := flag.String("cmd", "", "command: run | snapshot-save | snapshot-restore")
	file := flag.String("file", "", "snapshot file (for restore)")
	notes := flag.String("notes", "", "notes for snapshot")
	flag.Parse()

	switch *cmd {
	case "snapshot-save":
		path, err := storage.SaveSnapshot(".", map[string]any{"id": "NEON"}, []string{}, map[string]float64{}, map[string]bool{}, *notes)
		if err != nil {
			log.Fatalf("snapshot save failed: %v", err)
		}
		fmt.Println("Snapshot saved to", path)
		return

	case "snapshot-restore":
		if *file == "" {
			log.Fatal("must specify -file for snapshot-restore")
		}
		snap, err := storage.LoadSnapshot(*file)
		if err != nil {
			log.Fatalf("snapshot restore failed: %v", err)
		}
		fmt.Printf("Restored snapshot from %s (notes: %s)\n", snap.Timestamp, snap.Notes)
		return

	default:
		ctx := context.Background()
		agent := NewAgent()
		if err := agent.Run(ctx); err != nil {
			log.Fatalf("agent error: %v", err)
		}
	}
}
