package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
envs: []
env_file:
flags: []
cli: []

`

var multiGohotYml string = `# shared defaults (each app can override any of these)
ext:
 - .go
 - .yaml
ignore:
 - .git
 - vendor
debounce: 500

apps:
  api:
    path: ./services/api
    entry: ./services/api/main.go
    out: ./bin/api
  cli:
    path: ./cmd/cli
    entry: ./cmd/cli/main.go
    out: ./bin/cli
    cli: []
`

var __VERSION__ string = "1.3.0"

func loadConfigFile(readConfig bool) {
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
	found := findConfigFile()
	if found == "" {
		log.Fatalf("error: no gohot.yaml or gohot.yml found in this directory, its parents, or ~/.gohot/")
	}
	viper.SetConfigFile(found)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := validateApps(); err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Println("Loaded config:", viper.ConfigFileUsed())
}

func runFlags() []cli.Flag {
	return []cli.Flag{
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
	}
}

func runApp(c *cli.Context, app string) error {
	config := resolveConfig(c, app)
	if err := watcher.ValidateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	watcher.Start(config)
	return nil
}

func cwdConfigPath() string {
	for _, n := range configFileNames {
		if _, err := os.Stat(n); err == nil {
			fp, _ := filepath.Abs(n)
			return fp
		}
	}
	return ""
}

func main() {
	args, execName := extractExecName(os.Args)
	if len(args) < 2 {
		loadConfigFile(true)
	} else {
		loadConfigFile(args[1] != "init" && args[1] != "i" && args[1] != "--help" && args[1] != "-h" && args[1] != "version")
	}

	app := &cli.App{
		Name:                 "gohot",
		Usage:                "Auto-reload Go apps when source files change",
		EnableBashCompletion: true,
		Flags:                runFlags(),
		Commands: []*cli.Command{
			{
				Name:      "exec",
				Usage:     "run a named app from a multi-app gohot.yaml",
				ArgsUsage: "[name]",
				Flags:     runFlags(),
				BashComplete: func(c *cli.Context) {
					for _, n := range appNames() {
						fmt.Println(n)
					}
				},
				Action: func(c *cli.Context) error {
					name := execName
					if name == "" {
						name = c.Args().First()
					}
					sel, err := selectApp(appNames(), name)
					if err != nil {
						return err
					}
					return runApp(c, sel)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list apps defined in gohot.yaml",
				Action: func(c *cli.Context) error {
					names := appNames()
					if len(names) == 0 {
						fmt.Println("(single app; no `apps:` section)")
						return nil
					}
					for _, n := range names {
						path := viper.GetString(lookupKey(n, "path"))
						entry := viper.GetString(lookupKey(n, "entry"))
						fmt.Printf("%s\t%s\t%s\n", n, path, entry)
					}
					return nil
				},
			},
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "create default gohot.yaml file",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "multi", Usage: "scaffold a multi-app config with an apps: section"},
				},
				Action: func(ctx *cli.Context) error {
					if existing := cwdConfigPath(); existing != "" {
						fmt.Fprintf(os.Stderr, "File already exists: %s\n", existing)
						return nil
					}
					content := defaultGohotYml
					if ctx.Bool("multi") {
						content = multiGohotYml
					}
					if err := os.WriteFile("./gohot.yaml", []byte(content), 0o644); err != nil {
						return err
					}
					fmt.Fprintf(os.Stdout, "Created default config: %s\n", cwdConfigPath())
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
			sel, err := selectApp(appNames(), "")
			if err != nil {
				return err
			}
			return runApp(c, sel)
		},
	}
	if err := app.Run(args); err != nil {
		log.Fatal(err)
	}
}
