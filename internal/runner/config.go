package runner

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/axetroy/go-fs"
	"github.com/fatih/color"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type Action struct {
	Action    string
	Arguments string
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	CWD      string
	Actions  []Action
}

func NewRunner(configFile string) (*Config, error) {
	if fs.PathExists(configFile) == false {
		msg := fmt.Sprintf("Config file `%s` not found", configFile)
		return nil, errors.New(color.RedString(msg))
	}

	config, err := parseConfig(configFile)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *Config) Run() error {
	// ask password for remote server
	password := ""
	prompt := &survey.Password{
		Message: "Please type remote server's password",
	}

	if err := survey.AskOne(prompt, &password); err != nil {
		return err
	}

	r.Password = password

	client := NewSSH(*r)

	pwd, err := os.Getwd()

	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}

	defer client.Disconnect()

	for step, action := range r.Actions {

		switch action.Action {
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
				targetDir = path.Join(r.CWD, targetDir)
			}

			fmt.Printf("[Step %v]: COPY %s to %s\n", step+1, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(targetDir))

			for _, filePath := range sourceFiles {

				if path.IsAbs(filePath) == false {
					filePath = path.Join(pwd, filePath)
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

func parseConfig(configFile string) (c Config, err error) {
	var content []byte
	content, err = fs.ReadFile(configFile)

	raw := string(content[:])

	lines := strings.Split(raw, "\n")

	for _, line := range lines {
		s := strings.Trim(line, "")
		if s == "" {
			continue
		}

		arr := strings.Split(s, " ")

		keyword := arr[0]
		value := strings.Join(arr[1:], " ")

		switch keyword {
		case "HOST":
			c.Host = value
			break
		case "PORT":
			if port, e := strconv.Atoi(value); e != nil {
				err = e
				return
			} else {
				c.Port = port
			}
			break
		case "USERNAME":
			c.Username = value
			break
		case "CWD":
			c.CWD = value
			break
		case "COPY":
			c.Actions = append(c.Actions, Action{
				Action:    keyword,
				Arguments: value,
			})
			break
		case "RUN":
			c.Actions = append(c.Actions, Action{
				Action:    keyword,
				Arguments: value,
			})
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid keyword `%s`", color.RedString(keyword)))
			return
		}

	}

	return
}
