package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ableinc/gohot/watcher"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var appNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var configFileNames = []string{"gohot.yaml", "gohot.yml"}

// findConfigFileFrom looks for a config file in dir, then each parent
// directory, and finally in ~/.gohot/. Returns "" when none is found.
func findConfigFileFrom(dir string) string {
	for {
		for _, name := range configFileNames {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if home, err := os.UserHomeDir(); err == nil {
		for _, name := range configFileNames {
			p := filepath.Join(home, ".gohot", name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}
	return ""
}

func findConfigFile() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return findConfigFileFrom(wd)
}

// appNames returns the sorted app names defined under `apps:`.
func appNames() []string {
	apps := viper.GetStringMap("apps")
	names := make([]string, 0, len(apps))
	for n := range apps {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// validateApps checks app names and shapes. Duplicate output paths only warn.
func validateApps() error {
	apps := viper.GetStringMap("apps")
	outs := map[string]string{}
	for _, name := range appNames() {
		if !appNameRe.MatchString(name) {
			return fmt.Errorf("invalid app name %q (allowed: letters, digits, '-' and '_')", name)
		}
		switch apps[name].(type) {
		case nil, map[string]any:
		default:
			return fmt.Errorf("app %q must be a mapping of settings", name)
		}
		out := viper.GetString("out")
		if viper.IsSet("apps." + name + ".out") {
			out = viper.GetString("apps." + name + ".out")
		}
		out = filepath.Clean(out)
		if other, ok := outs[out]; ok {
			log.Printf("Warning: apps %q and %q share output path %s; set a distinct `out` for each", other, name, out)
		} else {
			outs[out] = name
		}
	}
	return nil
}

// selectApp decides which app to run. It returns "" for a legacy
// single-app config (no `apps:` section).
func selectApp(names []string, requested string) (string, error) {
	if len(names) == 0 {
		if requested != "" {
			return "", fmt.Errorf("no apps defined in config; cannot run %q", requested)
		}
		return "", nil
	}
	if requested == "" {
		if len(names) == 1 {
			return names[0], nil
		}
		return "", fmt.Errorf("multiple apps defined: %s. Use `gohot exec <name>`", strings.Join(names, ", "))
	}
	for _, n := range names {
		if strings.EqualFold(n, requested) {
			return n, nil
		}
	}
	return "", fmt.Errorf("unknown app %q. Available apps: %s", requested, strings.Join(names, ", "))
}

// lookupKey returns the viper key to read for a setting:
// the app block if it sets the key, otherwise the top-level key
// (which also carries the built-in defaults).
func lookupKey(app, key string) string {
	if app != "" {
		k := "apps." + app + "." + key
		if viper.IsSet(k) {
			return k
		}
	}
	return key
}

// resolveConfig builds the watcher config with precedence:
// CLI flag > app block > top-level > built-in default.
func resolveConfig(c *cli.Context, app string) watcher.Config {
	str := func(key string) string {
		if c.IsSet(key) {
			return c.String(key)
		}
		return viper.GetString(lookupKey(app, key))
	}
	strs := func(key string) []string {
		if c.IsSet(key) {
			return c.StringSlice(key)
		}
		return viper.GetStringSlice(lookupKey(app, key))
	}
	debounce := viper.GetInt(lookupKey(app, "debounce"))
	if c.IsSet("debounce") {
		debounce = c.Int("debounce")
	}
	return watcher.Config{
		Name:       app,
		Path:       str("path"),
		Extensions: strs("ext"),
		Ignore:     strs("ignore"),
		Output:     str("out"),
		MainFile:   str("entry"),
		Debounce:   time.Duration(debounce) * time.Millisecond,
		Envs:       strs("envs"),
		EnvFile:    str("env_file"),
		Flags:      strs("flags"),
		Cli:        strs("cli"),
	}
}

// extractExecName pulls the app name out of `gohot exec <name> ...` so that
// flags may appear before or after it (urfave/cli v2 stops parsing flags at
// the first positional argument).
func extractExecName(args []string) ([]string, string) {
	if len(args) >= 3 && args[1] == "exec" && !strings.HasPrefix(args[2], "-") {
		out := append([]string{}, args[:2]...)
		out = append(out, args[3:]...)
		return out, args[2]
	}
	return args, ""
}
