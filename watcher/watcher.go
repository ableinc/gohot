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
)

type Config struct {
	Path       string
	Extensions string
	Ignore     string
	Output     string
	MainFile   string
	Debounce   time.Duration
	Args       []string
}

var (
	cmd         *exec.Cmd
	cmdLock     sync.Mutex
	stdoutMulti io.Writer
	stderrMulti io.Writer
)

func splitExts(ext string) []string {
	var exts []string
	for e := range strings.SplitSeq(ext, ",") {
		exts = append(exts, strings.TrimSpace(e))
	}
	return exts
}

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

func runCommand(name string, goArgs []string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	setEnvs(goArgs, c)
	err := c.Start()
	if err != nil {
		log.Fatalf("error starting command: %v", err)
	}
	return c
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

func startProcess(cfg Config, mainFile string) {
	cmdLock.Lock()
	defer cmdLock.Unlock()

	if runtime.NumCPU() >= 4 {
		log.Println("Compiling binary...")
		var buildOutput bytes.Buffer
		stdoutMulti = io.MultiWriter(os.Stdout, &buildOutput)
		stderrMulti = io.MultiWriter(os.Stderr, &buildOutput)
		build := exec.Command("go", "build", "-o", cfg.Output, mainFile)
		build.Stdout = stdoutMulti
		build.Stderr = stderrMulti
		setEnvs(cfg.Args, build)
		err := build.Run()
		if err != nil {
			log.Println("Build failed. Falling back to go run...")
			log.Printf("Build output:\n%v\n", buildOutput.String())
			if strings.Contains(buildOutput.String(), "go : unknown command") {
				log.Println("Detected unknown command error in build output. Stopping process.")
				os.Exit(1)
			}
			cmd = runCommand("go", cfg.Args, "run", mainFile)
		} else {
			log.Println("Build succeeded. Running binary...")
			cmd = runCommand(cfg.Output, cfg.Args)
		}
	} else {
		log.Println("Low CPU system. Running with go run...")
		cmd = runCommand("go", cfg.Args, "run", mainFile)
	}
}

func Start(cfg Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	exts := splitExts(cfg.Extensions)
	ignores := splitExts(cfg.Ignore)
	mainFile := cfg.MainFile
	if mainFile == "" {
		mainFile, err = findMain(cfg.Path)
		if err != nil {
			log.Fatal("cannot locate main file:", err)
		}
	}
	err = addWatchers(watcher, cfg.Path, ignores)
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
			if isValidExt(event.Name, exts) && (event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Create != 0 || event.Op&fsnotify.Remove != 0) {
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
