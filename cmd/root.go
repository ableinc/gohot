package cmd

import (
	"log"
	"os"
	"time"

	"github.com/ableinc/gohot/watcher"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func loadConfigFile() {
	viper.SetConfigName("gohot")         // No extension
	viper.AddConfigPath(".")             // Look in current dir
	viper.AddConfigPath("$HOME/.gohot/") // Also support home config
	viper.SetConfigType("yaml")          // Default to YAML

	viper.SetDefault("path", "./")
	viper.SetDefault("ext", ".go")
	viper.SetDefault("out", "./appbin")
	viper.SetDefault("entry", "")
	viper.SetDefault("debounce", 300)

	err := viper.ReadInConfig()
	if err == nil {
		log.Println("Loaded config:", viper.ConfigFileUsed())
	}
}

func Execute() {
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
		Action: func(c *cli.Context) error {
			config := watcher.Config{
				Path:       c.String("path"),
				Extensions: c.String("ext"),
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
