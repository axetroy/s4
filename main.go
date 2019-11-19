package main

import (
	"log"
	"os"

	"github.com/axetroy/s4/core/command"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "s4"
	app.Usage = "Integrate local and remote workflow"
	app.Version = "0.8.1"
	app.Author = "Axetroy"
	app.Email = "axetroy.dev@gmail.com"

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
SOURCE CODE:
	https://github.com/axetroy/s4
`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "specify the s4 configuration file.",
			Value: ".s4", // default value
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "Print current s4 version to stdout",
			Action: func(c *cli.Context) error {
				return command.Version(app.Version)
			},
		},
		{
			Name:  "upgrade",
			Usage: "Upgrade s4 version to latest",
			Action: func(c *cli.Context) error {
				return command.Upgrade()
			},
		},
		{
			Name:  "init",
			Usage: "Initialize an s4 file",
			Action: func(c *cli.Context) error {
				return command.Init()
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		configFile := c.String("config")
		return command.Default(configFile)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
