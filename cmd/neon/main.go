package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"neon/internal/agent"
	"neon/internal/storage"
	"neon/internal/telemetry"
)

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
		logger := telemetry.NewLogger(".")
		defer logger.Close()

		ag := agent.NewAgent(logger)
		if err := ag.Run(ctx); err != nil {
			log.Fatalf("agent error: %v", err)
		}
	}
}
