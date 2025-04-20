package watcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Path       string
	Extensions string
	Output     string
	MainFile   string
	Debounce   time.Duration
}

var (
	cmd     *exec.Cmd
	cmdLock sync.Mutex
)

func splitExts(ext string) []string {
	var exts []string
	for _, e := range strings.Split(ext, ",") {
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

func addWatchers(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") && d.Name() != "vendor" {
			return watcher.Add(path)
		}
		return nil
	})
}

func runCommand(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
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

func startProcess(cfg Config, mainFile string) {
	cmdLock.Lock()
	defer cmdLock.Unlock()

	if runtime.NumCPU() >= 4 {
		log.Println("Compiling binary...")
		build := exec.Command("go", "build", "-o", cfg.Output, mainFile)
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		err := build.Run()
		if err != nil {
			log.Println("Build failed. Falling back to go run...")
			cmd = runCommand("go", "run", mainFile)
		} else {
			log.Println("Build succeeded. Running binary...")
			cmd = runCommand(cfg.Output)
		}
	} else {
		log.Println("Low CPU system. Running with go run...")
		cmd = runCommand("go", "run", mainFile)
	}
}

func Start(cfg Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	exts := splitExts(cfg.Extensions)
	mainFile := cfg.MainFile
	if mainFile == "" {
		mainFile, err = findMain(cfg.Path)
		if err != nil {
			log.Fatal("cannot locate main file:", err)
		}
	}
	err = addWatchers(watcher, cfg.Path)
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
