package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	watcher "github.com/ableinc/gohot/watcher"
)

func TestValidateConfig_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	tmpMain := filepath.Join(tmpDir, "main.go")
	os.WriteFile(tmpMain, []byte(`package main; func main(){}`), 0644)

	cfg := watcher.Config{
		Path:       tmpDir,
		Extensions: ".go,.yaml",
		MainFile:   tmpMain,
		Output:     filepath.Join(tmpDir, "appbin"),
		Debounce:   300 * time.Millisecond,
	}

	err := watcher.ValidateConfig(cfg)
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateConfig_InvalidCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name string
		cfg  watcher.Config
	}{
		{
			"missing path",
			watcher.Config{Path: "./does/not/exist", Extensions: ".go", Output: "appbin", Debounce: 100 * time.Millisecond},
		},
		{
			"bad extension format",
			watcher.Config{Path: tmpDir, Extensions: "go,yaml", Output: "appbin", Debounce: 100 * time.Millisecond},
		},
		{
			"non-positive debounce",
			watcher.Config{Path: tmpDir, Extensions: ".go", Output: "appbin", Debounce: 0},
		},
		{
			"main file doesn't exist",
			watcher.Config{Path: tmpDir, Extensions: ".go", MainFile: "fake.go", Output: "appbin", Debounce: 100 * time.Millisecond},
		},
		{
			"main file not .go",
			watcher.Config{Path: tmpDir, Extensions: ".go", MainFile: "main.txt", Output: "appbin", Debounce: 100 * time.Millisecond},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.ValidateConfig(tt.cfg)
			if err == nil {
				t.Fatalf("expected error for case: %s", tt.name)
			}
		})
	}
}
