package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ableinc/gohot/watcher"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var defaultGohotYml string = `path: "./"
ext:
 - .go
 - .yaml
ignore:
 - .git
 - vendor
out: ./appb
entry: main.go
debounce: 500
envs:
 -
env_file:
flags:
 -
cli:
 -

`

var __VERSION__ string = "1.0.1"

func loadConfigFile(readConfig bool) {
	viper.SetConfigName("gohot")         // No extension
	viper.AddConfigPath(".")             // Look in current dir
	viper.AddConfigPath("$HOME/.gohot/") // Also support home config
	viper.SetConfigType("yaml")

	viper.SetDefault("path", "./")
	viper.SetDefault("ext", []string{".go", ".yaml"})
	viper.SetDefault("ignore", []string{".git", "vendor"})
	viper.SetDefault("out", "./appb")
	viper.SetDefault("entry", "main.go")
	viper.SetDefault("debounce", 1000)
	viper.SetDefault("envs", []string{""})
	viper.SetDefault("env_file", "")
	viper.SetDefault("flags", []string{""})
	viper.SetDefault("cli", []string{""})
	if !readConfig {
		return
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Println("Loaded config:", viper.ConfigFileUsed())
}

func getConfigFilePath() string {
	if _, err := os.Stat("./gohot.yaml"); err == nil {
		fp, _ := filepath.Abs("./gohot.yaml")
		return fp
	}
	if _, err := os.Stat("./gohot.yml"); err == nil {
		fp, _ := filepath.Abs("./gohot.yml")
		return fp
	}
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(path.Join(home, ".gohot", "gohot.yaml")); err == nil {
			fp, _ := filepath.Abs(path.Join(home, ".gohot", "gohot.yaml"))
			return fp
		}
		if _, err := os.Stat(path.Join(home, ".gohot", "gohot.yml")); err == nil {
			fp, _ := filepath.Abs(path.Join(home, ".gohot", "gohot.yml"))
			return fp
		}
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		loadConfigFile(true)
	} else {
		loadConfigFile(os.Args[1] != "init" && os.Args[1] != "i" && os.Args[1] != "--help" && os.Args[1] != "-h" && os.Args[1] != "version")
	}

	app := &cli.App{
		Name:  "gohot",
		Usage: "Auto-reload Go apps when source files change",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Directory to watch",
				Value:   viper.GetString("path"),
			},
			&cli.StringSliceFlag{
				Name:    "ext",
				Aliases: []string{"e"},
				Usage:   "File extension to watch",
				Value:   cli.NewStringSlice(viper.GetStringSlice("ext")...),
			},
			&cli.StringSliceFlag{
				Name:    "ignore",
				Aliases: []string{"i"},
				Usage:   "File paths to ignore",
				Value:   cli.NewStringSlice(viper.GetStringSlice("ignore")...),
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage:   "Output binary name when compiling",
				Value:   viper.GetString("out"),
			},
			&cli.StringFlag{
				Name:    "entry",
				Aliases: []string{"m"},
				Usage:   "Main Go file entry point",
				Value:   viper.GetString("entry"),
			},
			&cli.IntFlag{
				Name:    "debounce",
				Aliases: []string{"d"},
				Usage:   "Debounce time in milliseconds",
				Value:   viper.GetInt("debounce"),
			},
			&cli.StringSliceFlag{
				Name:    "envs",
				Aliases: []string{"v"},
				Usage:   "Environment variables to set before go build or go run",
				Value:   cli.NewStringSlice(viper.GetStringSlice("envs")...),
			},
			&cli.StringFlag{
				Name:    "env_file",
				Aliases: []string{"env"},
				Usage:   "Path to .env file to load environment variables from",
				Value:   viper.GetString("env_file"),
			},
			&cli.StringSliceFlag{
				Name:    "flags",
				Aliases: []string{"f"},
				Usage:   "Build flags to pass to go build",
				Value:   cli.NewStringSlice(viper.GetStringSlice("flags")...),
			},
			&cli.StringSliceFlag{
				Name:    "cli",
				Aliases: []string{"c"},
				Usage:   "CLI arguments to pass to the compiled binary",
				Value:   cli.NewStringSlice(viper.GetStringSlice("cli")...),
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "create default gohot.yaml file",
				Action: func(ctx *cli.Context) error {
					_, ymlErr := os.Stat("./gohot.yml")
					_, err := os.Stat("./gohot.yaml")
					if err != nil && ymlErr != nil {
						err = os.WriteFile("./gohot.yaml", []byte(defaultGohotYml), 0644)
						if err != nil {
							return err
						}
						fmt.Fprintf(os.Stdout, "Created default config: %s\n", getConfigFilePath())
						return nil
					}
					fmt.Fprintf(os.Stderr, "File already exists: %s\n", getConfigFilePath())
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Print the version number",
				Action: func(c *cli.Context) error {
					fmt.Printf("gohot version %s\n", __VERSION__)
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			config := watcher.Config{
				Path:       c.String("path"),
				Extensions: c.StringSlice("ext"),
				Ignore:     c.StringSlice("ignore"),
				Output:     c.String("out"),
				MainFile:   c.String("entry"),
				Debounce:   time.Duration(c.Int("debounce")) * time.Millisecond,
				Envs:       c.StringSlice("envs"),
				EnvFile:    c.String("env_file"),
				Flags:      c.StringSlice("flags"),
				Cli:        c.StringSlice("cli"),
			}

			if err := watcher.ValidateConfig(config); err != nil {
				log.Fatalf("Invalid configuration: %v", err)
			}

			watcher.Start(config)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
