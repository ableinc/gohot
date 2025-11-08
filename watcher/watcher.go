package watcher

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/shlex"
)

type Config struct {
	Path       string
	Extensions []string
	Ignore     []string
	Output     string
	MainFile   string
	Debounce   time.Duration
	Envs       []string
	Flags      []string
	Cli        []string
}

var (
	cmd         *exec.Cmd
	cmdLock     sync.Mutex
	stdoutMulti io.Writer
	stderrMulti io.Writer
)

func isValidExt(file string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

func findMain(path string) (string, error) {
	mainFile := filepath.Join(path, "main.go")
	if _, err := os.Stat(mainFile); err == nil {
		return mainFile, nil
	}
	return "", fmt.Errorf("main.go not found in %s", path)
}

func parseFlag(flagsStr string) []string {
	args, err := shlex.Split(flagsStr)
	if err != nil {
		log.Printf("Error parsing flags: %v", err)
		return []string{}
	}
	return args
}

func addWatchers(watcher *fsnotify.Watcher, root string, ignores []string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if current path should be ignored
		if slices.Contains(ignores, path) || slices.Contains(ignores, d.Name()) {
			if d.IsDir() {
				// Skip the entire directory and its contents
				return filepath.SkipDir
			}
			// Skip just this file, continue with directory traversal
			return nil
		}

		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			return watcher.Add(path)
		}
		return nil
	})
}

func executeCommand(name string, cfg Config, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	setEnvs(cfg.Envs, c)
	err := c.Start()
	if err != nil {
		log.Fatalf("error starting command: %v", err)
	}
	return c
}

func buildCommand(cfg Config, mainFile string) (*exec.Cmd, bytes.Buffer) {
	var buildOutput bytes.Buffer
	stdoutMulti = io.MultiWriter(os.Stdout, &buildOutput)
	stderrMulti = io.MultiWriter(os.Stderr, &buildOutput)
	build := exec.Command("go", buildCommandArgs(cfg, mainFile)...)
	build.Stdout = stdoutMulti
	build.Stderr = stderrMulti
	// Set environment variables at the build stage as well
	setEnvs(cfg.Envs, build)
	return build, buildOutput
}

func stopProcess() {
	cmdLock.Lock()
	defer cmdLock.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		cmd.Wait()
	}
}

func setEnvs(goArgs []string, cmd *exec.Cmd) {
	cmd.Env = append(os.Environ(), goArgs...)
}

func buildCommandArgs(cfg Config, mainFile string) []string {
	args := []string{"build"}
	if len(cfg.Flags) > 0 {
		// Sanitize and parse flags
		for i := range cfg.Flags {
			flag := strings.TrimSpace(cfg.Flags[i])
			if !strings.HasPrefix(flag, "-") {
				flag = "-" + flag
			}
			args = append(args, parseFlag(flag)...)
		}
	}
	args = append(args, "-o", cfg.Output, mainFile)
	return args
}

func appendCliArgs(cfg Config, args []string) []string {
	for i := range cfg.Cli {
		cli := strings.TrimSpace(cfg.Cli[i])
		args = append(args, cli)
	}
	return args
}

func runCommandArgs(cfg Config, mainFile string) []string {
	args := []string{"run"}
	args = append(args, mainFile)
	// Append CLI arguments if provided
	args = appendCliArgs(cfg, args)
	return args
}

func startProcess(cfg Config, mainFile string) {
	cmdLock.Lock()
	defer cmdLock.Unlock()

	if runtime.NumCPU() >= 4 {
		log.Println("Compiling binary...")
		build, buildOutput := buildCommand(cfg, mainFile)
		err := build.Run()
		if err != nil {
			log.Println("Build failed. Falling back to go run...")
			log.Printf("Build output:\n%v\n", buildOutput.String())
			if strings.Contains(buildOutput.String(), "go : unknown command") {
				log.Println("Detected unknown command error in build output. Stopping process.")
				os.Exit(1)
			}
			cmd = executeCommand("go", cfg, runCommandArgs(cfg, mainFile)...)
		} else {
			log.Println("Build succeeded. Running binary...")
			cmd = executeCommand(cfg.Output, cfg, appendCliArgs(cfg, []string{})...)
		}
	} else {
		log.Println("Low CPU system. Running with go run...")
		cmd = executeCommand("go", cfg, runCommandArgs(cfg, mainFile)...)
	}
}

func Start(cfg Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	mainFile := cfg.MainFile
	if mainFile == "" {
		mainFile, err = findMain(cfg.Path)
		if err != nil {
			log.Fatal("cannot locate main file:", err)
		}
	}
	err = addWatchers(watcher, cfg.Path, cfg.Ignore)
	if err != nil {
		log.Fatal(err)
	}
	startProcess(cfg, mainFile)
	var debouncerTimer *time.Timer
	changed := false
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if isValidExt(event.Name, cfg.Extensions) && (event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Create != 0 || event.Op&fsnotify.Remove != 0) {
				log.Println("File changed:", event.Name)
				changed = true
				if debouncerTimer != nil {
					debouncerTimer.Stop()
				}
				debouncerTimer = time.AfterFunc(cfg.Debounce, func() {
					if changed {
						stopProcess()
						startProcess(cfg, mainFile)
						changed = false
					}
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Fatal("watcher error:", err)
		}
	}
}
