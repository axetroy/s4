package runner

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner/actions"
	"github.com/axetroy/s4/src/ssh"
	"github.com/fatih/color"
	"os"
)

type Runner struct {
	SSH    *ssh.Client
	Config *configuration.Configuration
	Step   int    // current step
	Cwd    string // current working dir at local
}

func NewRunner(configFilepath string) (*Runner, error) {
	if f, err := os.Stat(configFilepath); err != nil {
		msg := fmt.Sprintf("Config file `%s` not found", configFilepath)
		return nil, errors.New(color.RedString(msg))
	} else {
		if f.IsDir() {
			msg := fmt.Sprintf("Config file `%s` is not a file", configFilepath)
			return nil, errors.New(color.RedString(msg))
		}
	}

	fmt.Printf("Load the s4 file `%s`\n", color.GreenString(configFilepath))

	config, err := configuration.ParseFile(configFilepath)

	if err != nil {
		return nil, err
	}

	return &Runner{
		Config: config,
		Step:   1,
	}, nil
}

func (r *Runner) Run() error {
	client := ssh.NewSSH(r.Config)
	r.SSH = client

	r.Step++

	fmt.Printf("[step %v]: CONNECT %s\n", r.Step, color.GreenString(fmt.Sprintf("%s@%s:%s", r.Config.Username, r.Config.Host, r.Config.Port)))

	if r.Config.Password == "" {
		// ask password for remote server
		password := ""
		prompt := &survey.Password{
			Message: "Please type remote server's password",
		}

		if err := survey.AskOne(prompt, &password); err != nil {
			return err
		}

		r.Config.Password = password
	}

	if err := client.Connect(); err != nil {
		return err
	}

	defer client.Disconnect()

	if cwd, err := os.Getwd(); err != nil {
		return err
	} else {
		r.Cwd = cwd
	}

	remoteCwd, err := client.Pwd()

	if err != nil {
		return err
	}

	r.Config.CWD = remoteCwd

	for _, action := range r.Config.Actions {
		r.Step++
		switch action.Action {
		case "VAR":
			if err := actions.Var(r, action); err != nil {
				return err
			}
			break
		case "CD":
			if err := actions.Cd(r, action); err != nil {
				return err
			}
			break
		case "BASH":
			if err := actions.Bash(r, action); err != nil {
				return err
			}
			break
		case "CMD":
			if err := actions.Cmd(r, action); err != nil {
				return err
			}
			break
		case "RUN":
			if err := actions.Run(r, action); err != nil {
				return err
			}
			break
		case "MOVE":
			if err := actions.Move(r, action); err != nil {
				return err
			}
			break
		case "COPY":
			if err := actions.Copy(r, action); err != nil {
				return err
			}
			break
		case "DELETE":
			if err := actions.Delete(r, action); err != nil {
				return err
			}
			break
		case "UPLOAD":
			if err := actions.Upload(r, action); err != nil {
				return err
			}
			break
		case "DOWNLOAD":
			if err := actions.Download(r, action); err != nil {
				return err
			}
			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Action))
		}
	}

	return nil
}
