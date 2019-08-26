package main

import (
	"fmt"
	"github.com/axetroy/sshunter/lib/runner"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Name = "s4"
	app.Usage = "Perform remote server tasks on local computer"
	app.Version = "0.1.0"

	cli.AppHelpTemplate = fmt.Sprintf(`%s

WEBSITE: https://github.com/axetroy/s4

REPORT BUGS: https://github.com/axetroy/s4/issues

`, cli.AppHelpTemplate)

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

		if password != "" {
			r.Config.Password = password
		}

		if err != nil {
			return err
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
