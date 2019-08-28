package lib

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Runner struct {
	Config *Config
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

	config, err := ParseFile(configFilepath)

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

	client := NewSSH(r.Config)

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

	step := 1

	for _, action := range r.Config.Actions {
		argument := strings.Join(action.Arguments, " ")

		switch action.Action {
		case "CD":
			dir := argument
			r.Config.CWD = dir
			fmt.Printf("[Step %v]: CD %s\n", step, color.GreenString(dir))
			step += 1
			break
		case "BASH":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", action.Arguments))
			fmt.Printf("[Step %v]: CMD %s\n", step, commandWithColor)

			bashPath := os.Getenv("SHELL")

			// if not found bash in you env.
			if len(bashPath) == 0 {
				if bashBinPath, bashNotExist := exec.LookPath("bash"); bashNotExist != nil {
					if shBinPath, shNotExist := exec.LookPath("sh"); shNotExist != nil {
						return errors.New(" can not found 'bash' in your system")
					} else {
						bashPath = shBinPath
					}
				} else {
					bashPath = bashBinPath
				}
			}

			c := exec.Command(bashPath, "-c", argument)

			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				return err
			}

			if c.ProcessState.Success() == false {
				return errors.New(fmt.Sprintf("run command '%s' fail.", action.Arguments))
			}

			step += 1

			break
		case "CMD":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", action.Arguments))

			fmt.Printf("[Step %v]: CMD %s\n", step, commandWithColor)

			command := action.Arguments[0]
			args := action.Arguments[1:]

			if _, err := exec.LookPath(command); err != nil {
				fmt.Printf("didn't find '%s' executable\n", command)
			}

			c := exec.Command(command, args...)

			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				return err
			}

			if c.ProcessState.Success() == false {
				return errors.New(fmt.Sprintf("run command '%s' fail.", action.Arguments))
			}

			step += 1
			break
		case "RUN":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", action.Arguments))

			fmt.Printf("[Step %v]: RUN %s\n", step, commandWithColor)

			if err := client.Run(argument); err != nil {
				return err
			}

			step += 1
			break
		case "MOVE":
			args := strings.Split(argument, " ")

			if len(args) != 2 {
				return errors.New(fmt.Sprintf("move require source and destination but got `%s`", args))
			}

			sourceFilepath := strings.Trim(args[0], " ")
			destinationFilepath := strings.Trim(args[1], " ")

			if path.IsAbs(sourceFilepath) == false {
				sourceFilepath = path.Join(r.Config.CWD, sourceFilepath)
			}

			if path.IsAbs(destinationFilepath) == false {
				destinationFilepath = path.Join(r.Config.CWD, destinationFilepath)
			}

			fmt.Printf("[Step %v]: MOVE %s to %s\n", step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

			if err := client.Move(sourceFilepath, destinationFilepath); err != nil {
				return err
			}

			step += 1

			break
		case "COPY":
			args := action.Arguments

			if len(args) != 2 {
				return errors.New(fmt.Sprintf("copy require source and destination but got `%s`", args))
			}

			sourceFilepath := strings.Trim(args[0], " ")
			destinationFilepath := strings.Trim(args[1], " ")

			if path.IsAbs(sourceFilepath) == false {
				sourceFilepath = path.Join(r.Config.CWD, sourceFilepath)
			}

			if path.IsAbs(destinationFilepath) == false {
				destinationFilepath = path.Join(r.Config.CWD, destinationFilepath)
			}

			fmt.Printf("[Step %v]: COPY %s to %s\n", step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

			if err := client.Copy(sourceFilepath, destinationFilepath); err != nil {
				return err
			}

			step += 1

			break
		case "DELETE":
			args := action.Arguments

			var files []string

			for _, file := range args {
				if path.IsAbs(file) == false {
					file = path.Join(r.Config.CWD, file)
				}

				files = append(files, file)
			}

			fmt.Printf("[Step %v]: DELETE %s\n", step, color.YellowString(argument))

			if err := client.Delete(files...); err != nil {
				return err
			}

			step += 1

			break
		case "UPLOAD":
			f, err := FileParser(argument)

			if err != nil {
				return err
			}

			if path.IsAbs(f.Destination) == false {
				if r.Config.CWD != "" {
					f.Destination = path.Join(r.Config.CWD, f.Destination)
				}
			}

			fmt.Printf("[Step %v]: UPLOAD local:%s to remote:%s\n", step, color.YellowString(strings.Join(f.Source, ", ")), color.GreenString(f.Destination))

			for _, filePath := range f.Source {

				if path.IsAbs(filePath) == false {
					filePath = path.Join(localCwd, filePath)
				}

				err := client.Upload(filePath, f.Destination)

				if err != nil {
					return err
				}
			}

			step += 1

			break
		case "DOWNLOAD":
			f, err := FileParser(argument)

			if err != nil {
				return err
			}

			if path.IsAbs(f.Destination) == false {
				f.Destination = path.Join(localCwd, f.Destination)
			}

			fmt.Printf("[Step %v]: DOWNLOAD remote:%s to local:%s\n", step, color.YellowString(strings.Join(f.Source, ", ")), color.GreenString(f.Destination))

			for _, filePath := range f.Source {

				if path.IsAbs(filePath) == false {
					if r.Config.CWD != "" {
						filePath = path.Join(r.Config.CWD, filePath)
					}
				}

				err := client.Download(filePath, f.Destination)

				fmt.Println("download", filePath, "-->", f.Destination)

				if err != nil {
					return err
				}
			}

			step += 1

			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Action))
		}
	}

	return nil
}
