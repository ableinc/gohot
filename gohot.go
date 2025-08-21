package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ableinc/gohot/watcher"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var defaultGohotYml string = `
path: "./"
ext: .go,.yaml
ignore: .git,vendor
out: ./appb
entry: ./main.go
debounce: 500
`

func loadConfigFile() {
	viper.SetConfigName("gohot")         // No extension
	viper.AddConfigPath(".")             // Look in current dir
	viper.AddConfigPath("$HOME/.gohot/") // Also support home config
	viper.SetConfigType("yaml")          // Default to YAML

	viper.SetDefault("path", "./")
	viper.SetDefault("ext", ".go")
	viper.SetDefault("ignore", ".git,vendor")
	viper.SetDefault("out", "./appb")
	viper.SetDefault("entry", "main.go")
	viper.SetDefault("debounce", 500)

	err := viper.ReadInConfig()
	if err == nil {
		log.Println("Loaded config:", viper.ConfigFileUsed())
	}
}

func main() {
	loadConfigFile()

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
			&cli.StringFlag{
				Name:    "ext",
				Aliases: []string{"e"},
				Usage:   "File extension to watch (comma-separated)",
				Value:   viper.GetString("ext"),
			},
			&cli.StringFlag{
				Name:  "ignore",
				Usage: "File paths to ignore (comma-separated)",
				Value: viper.GetString("ignore"),
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
		},
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "create default gohot.yaml file",
				Action: func(ctx *cli.Context) error {
					_, err := os.Stat("./gohot.yaml")
					if err != nil {
						err = os.WriteFile("./gohot.yaml", []byte(defaultGohotYml), 0644)
						if err != nil {
							return err
						}
						return nil
					}
					fmt.Fprintf(os.Stderr, "File already exists: %s\n", viper.ConfigFileUsed())
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			config := watcher.Config{
				Path:       c.String("path"),
				Extensions: c.String("ext"),
				Ignore:     c.String("ignore"),
				Output:     c.String("out"),
				MainFile:   c.String("entry"),
				Debounce:   time.Duration(c.Int("debounce")) * time.Millisecond,
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
