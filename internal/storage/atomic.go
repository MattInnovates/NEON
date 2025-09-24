package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Write JSON atomically: write -> fsync -> rename
func AtomicWriteJSON(path string, v any) error {
	tmp := path + ".tmp"

	// open temp file
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		_ = f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func ReadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
