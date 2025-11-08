package watcher

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func ValidateConfig(cfg Config) error {
	// Validate watched path
	info, err := os.Stat(cfg.Path)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid watch path: %s", cfg.Path)
	}
	// Validate extensions
	if len(cfg.Extensions) == 0 {
		return errors.New("no file extensions specified")
	}
	for i := range cfg.Extensions {
		if !strings.HasPrefix(strings.TrimSpace(cfg.Extensions[i]), ".") {
			return fmt.Errorf("invalid extension format: %s (must start with a dot)", cfg.Extensions[i])
		}
	}

	// Validate debounce
	if cfg.Debounce <= 0 {
		return fmt.Errorf("debounce must be > 0")
	}

	// Validate output
	if cfg.Output == "" {
		return errors.New("output binary path is required")
	}
	if stat, err := os.Stat(cfg.Output); err == nil && stat.IsDir() {
		return fmt.Errorf("output path '%s' is a directory", cfg.Output)
	}

	// Validate main entry file if provided
	if cfg.MainFile != "" {
		if _, err := os.Stat(cfg.MainFile); err != nil {
			return fmt.Errorf("main file does not exist: %s", cfg.MainFile)
		}
		if !strings.HasSuffix(cfg.MainFile, ".go") {
			return fmt.Errorf("main file must be a .go file: %s", cfg.MainFile)
		}
	}

	return nil
}
