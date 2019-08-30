package main

import (
	"github.com/axetroy/s4/src/command"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Name = "s4"
	app.Usage = "Integrate local and remote workflow"
	app.Version = "0.4.0"
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
			Usage: "The s4 configuration file.",
			Value: ".s4", // default value
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "Specify the password for the server",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "print current s4 version",
			Action: func(c *cli.Context) error {
				return command.Version(app.Version)
			},
		},
		{
			Name:  "upgrade",
			Usage: "upgrade s4 version to latest",
			Action: func(c *cli.Context) error {
				return command.Upgrade()
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		configFile := c.String("config")
		password := c.String("password")
		return command.Detault(configFile, password)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
