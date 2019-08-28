package main

import (
	"github.com/axetroy/s4/lib/runner"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Name = "s4"
	app.Usage = "Perform remote server tasks on local computer"
	app.Version = "0.1.3"
	app.Author = "Axetroy"
	app.Email = "axetroy.dev@gmail.com"

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
WEBSITE: https://github.com/axetroy/s4
REPORT BUGS: https://github.com/axetroy/s4/issues
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

	app.Action = func(c *cli.Context) error {
		configFile := c.String("config")
		password := c.String("password")

		r, err := runner.NewRunner(configFile)

		if err != nil {
			return err
		}

		if password != "" {
			r.Config.Password = password
		}

		if err := r.Run(); err != nil {
			return err
		}

		return nil
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}

}
