package runner

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/axetroy/go-fs"
	"github.com/axetroy/sshunter/lib/parser"
	"github.com/fatih/color"
	"os"
	"path"
	"regexp"
	"strings"
)

type Runner struct {
	Config *parser.Config
}

func NewRunner(configFile string) (*Runner, error) {
	if fs.PathExists(configFile) == false {
		msg := fmt.Sprintf("Config file `%s` not found", configFile)
		return nil, errors.New(color.RedString(msg))
	}

	config, err := parser.ParseFile(configFile)

	if err != nil {
		return nil, err
	}

	return &Runner{
		Config: config,
	}, nil
}

func (r *Runner) Run() error {
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

	client := NewSSH(*r.Config)

	localCwd, err := os.Getwd()

	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}

	defer client.Disconnect()

	remoteCwd, err := client.Pwd()

	if err != nil {
		return err
	}

	r.Config.CWD = remoteCwd

	for step, action := range r.Config.Actions {

		switch action.Action {
		case "CWD":
			r.Config.CWD = action.Arguments
			fmt.Printf("[Step %v]: CWD %s\n", step+1, color.GreenString(action.Arguments))
			break
		case "RUN":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", action.Arguments))

			fmt.Printf("[Step %v]: RUN %s\n", step+1, commandWithColor)

			if err := client.Run(action.Arguments); err != nil {
				return err
			}
			break
		case "COPY":
			files := regexp.MustCompile("\\s+").Split(action.Arguments, -1)

			if len(files) < 2 {
				return errors.New("invalid Copy command")
			}

			lastElementIndex := len(files)

			sourceFiles := files[:lastElementIndex-1]
			targetDir := files[lastElementIndex-1]

			if path.IsAbs(targetDir) == false {
				if r.Config.CWD != "" {
					targetDir = path.Join(r.Config.CWD, targetDir)
				}
			}

			fmt.Printf("[Step %v]: COPY local:%s to remote:%s\n", step+1, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(targetDir))

			for _, filePath := range sourceFiles {

				if path.IsAbs(filePath) == false {
					filePath = path.Join(localCwd, filePath)
				}

				err := client.Copy(filePath, targetDir)

				fmt.Println("copy", filePath, "-->", targetDir)

				if err != nil {
					return err
				}
			}

			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Action))
		}
	}

	return nil
}
