package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const multiYAML = `
ext: [.go]
debounce: 700
ignore: [.git]
apps:
  api:
    path: ./services/api
    out: ./bin/api
    debounce: 200
  cli:
    path: ./cmd/cli
`

func loadYAML(t *testing.T, y string) {
	t.Helper()
	viper.Reset()
	loadConfigFileDefaultsOnly()
	dir := t.TempDir()
	p := filepath.Join(dir, "gohot.yaml")
	if err := os.WriteFile(p, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(p)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
}

func loadConfigFileDefaultsOnly() { loadConfigFile(false) }

func TestSelectApp(t *testing.T) {
	tests := []struct {
		name      string
		names     []string
		requested string
		want      string
		wantErr   bool
	}{
		{"legacy no name", nil, "", "", false},
		{"legacy with name", nil, "api", "", true},
		{"single app implicit", []string{"api"}, "", "api", false},
		{"multiple no name", []string{"api", "cli"}, "", "", true},
		{"unknown", []string{"api", "cli"}, "nope", "", true},
		{"case insensitive", []string{"api", "cli"}, "CLI", "cli", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectApp(tt.names, tt.requested)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("got (%q, %v), want (%q, err=%v)", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestAppNamesAndPrecedence(t *testing.T) {
	loadYAML(t, multiYAML)
	if got := appNames(); !reflect.DeepEqual(got, []string{"api", "cli"}) {
		t.Fatalf("appNames = %v", got)
	}
	// app block beats top-level
	if k := lookupKey("api", "debounce"); k != "apps.api.debounce" {
		t.Fatalf("api debounce key = %s", k)
	}
	// top-level beats built-in default
	if k := lookupKey("cli", "debounce"); k != "debounce" {
		t.Fatalf("cli debounce key = %s", k)
	}
	if viper.GetInt("debounce") != 700 {
		t.Fatalf("top-level debounce = %d", viper.GetInt("debounce"))
	}
	// built-in default when unset anywhere
	if got := viper.GetString(lookupKey("cli", "entry")); got != "main.go" {
		t.Fatalf("default entry = %q", got)
	}
}

func runResolve(t *testing.T, app string, args ...string) (debounce int64, path string) {
	t.Helper()
	a := &cli.App{Flags: runFlags(), Action: func(c *cli.Context) error {
		cfg := resolveConfig(c, app)
		debounce = cfg.Debounce.Milliseconds()
		path = cfg.Path
		return nil
	}}
	if err := a.Run(append([]string{"gohot"}, args...)); err != nil {
		t.Fatal(err)
	}
	return
}

func TestResolveConfigPrecedence(t *testing.T) {
	loadYAML(t, multiYAML)
	if d, p := runResolve(t, "api"); d != 200 || p != "./services/api" {
		t.Fatalf("api: debounce=%d path=%s", d, p)
	}
	if d, _ := runResolve(t, "cli"); d != 700 {
		t.Fatalf("cli debounce = %d, want 700", d)
	}
	if d, _ := runResolve(t, "api", "--debounce", "50"); d != 50 {
		t.Fatalf("flag override debounce = %d, want 50", d)
	}
}

func TestValidateApps(t *testing.T) {
	loadYAML(t, "apps:\n  \"bad.name\":\n    path: ./x\n")
	if err := validateApps(); err == nil {
		t.Fatal("expected invalid name error")
	}
	loadYAML(t, "apps:\n  a:\n    path: ./a\n  b:\n    path: ./b\n")
	if err := validateApps(); err != nil { // duplicate out only warns
		t.Fatalf("unexpected error: %v", err)
	}
	loadYAML(t, "apps:\n  a: 5\n")
	if err := validateApps(); err == nil {
		t.Fatal("expected non-mapping error")
	}
}

func TestFindConfigFileWalksUp(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "gohot.yml")
	if err := os.WriteFile(p, []byte("path: ./\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findConfigFileFrom(nested); got != p {
		t.Fatalf("got %q want %q", got, p)
	}
}

func TestExtractExecName(t *testing.T) {
	args, name := extractExecName([]string{"gohot", "exec", "api", "--debounce", "5"})
	if name != "api" || !reflect.DeepEqual(args, []string{"gohot", "exec", "--debounce", "5"}) {
		t.Fatalf("got %v %q", args, name)
	}
	args, name = extractExecName([]string{"gohot", "exec", "--debounce", "5", "api"})
	if name != "" || len(args) != 5 {
		t.Fatalf("got %v %q", args, name)
	}
}
