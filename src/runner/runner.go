package runner

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/ssh"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Runner struct {
	Config *configuration.Configuration
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
	}, nil
}

func (r *Runner) Run() error {
	client := ssh.NewSSH(r.Config)

	step := 1

	fmt.Printf("[Step %v]: CONNECT %s\n", step, color.GreenString(fmt.Sprintf("%s@%s:%s", r.Config.Username, r.Config.Host, r.Config.Port)))

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

	step++

	defer client.Disconnect()

	localCwd, err := os.Getwd()

	if err != nil {
		return err
	}

	remoteCwd, err := client.Pwd()

	if err != nil {
		return err
	}

	r.Config.CWD = remoteCwd

	for _, action := range r.Config.Actions {
		argument := strings.Join(action.Arguments, " ")

		switch action.Action {
		case "VAR":
			fmt.Printf("[Step %v]: VAR %s\n", step, color.GreenString(argument))
			step++

			if Var, err := variable.Parse(argument); err != nil {
				return err
			} else {
				switch Var.Type {
				case variable.TypeLiteral:
					r.Config.Var[Var.Key] = Var.Value
					break
				case variable.TypeEnv:
					if Var.Remote == false {
						// get local env
						r.Config.Var[Var.Key] = os.Getenv(Var.Value)
					} else {
						// get remote env
						remoteEnvValue, err := client.Env(Var.Value)

						if err != nil {
							return err
						}

						r.Config.Var[Var.Key] = remoteEnvValue
					}
					break
				case variable.TypeCommand:
					if Var.Remote == false {
						// execute command at local

						arr := strings.Split(Var.Value, " ")
						command := arr[0]
						args := arr[1:]

						c := exec.Command(command, args...)

						var stdoutBuf bytes.Buffer
						var stderrBuf bytes.Buffer

						c.Stdout = &stdoutBuf
						c.Stderr = &stderrBuf

						if err := c.Run(); err != nil {
							return err
						}

						if c.ProcessState.Success() == false {
							return errors.New(fmt.Sprintf("run command '%s' fail.", action.Arguments))
						}

						r.Config.Var[Var.Key] = strings.TrimSpace(stdoutBuf.String())
					} else {
						// execute command at remote
						var stdoutBuf bytes.Buffer
						var stderrBuf bytes.Buffer

						err := client.RunRaw(Var.Value, &stdoutBuf, &stderrBuf)

						if err != nil {
							return err
						}

						r.Config.Var[Var.Key] = strings.TrimSpace(stdoutBuf.String())

						fmt.Println(r.Config.Var[Var.Key])
					}
					break
				}
			}
			break
		case "CD":
			dir := argument
			r.Config.CWD = variable.Compile(dir, r.Config.Var)
			fmt.Printf("[Step %v]: CD %s\n", step, color.GreenString(dir))
			step++
			break
		case "BASH":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", argument))
			fmt.Printf("[Step %v]: BASH %s\n", step, commandWithColor)

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

			command := variable.Compile(argument, r.Config.Var)

			c := exec.Command(bashPath, "-c", command)

			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				return err
			}

			if c.ProcessState.Success() == false {
				return errors.New(fmt.Sprintf("run command '%s' fail.", action.Arguments))
			}

			step++

			break
		case "CMD":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", action.Arguments))

			fmt.Printf("[Step %v]: CMD %s\n", step, commandWithColor)

			command := variable.Compile(action.Arguments[0], r.Config.Var)
			args := variable.CompileArray(action.Arguments[1:], r.Config.Var)

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

			step++
			break
		case "RUN":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", argument))

			fmt.Printf("[Step %v]: RUN %s\n", step, commandWithColor)

			command := variable.Compile(argument, r.Config.Var)

			if err := client.Run(command); err != nil {
				return err
			}

			step++
			break
		case "MOVE":
			sourceFilepath := variable.Compile(action.Arguments[0], r.Config.Var)
			destinationFilepath := variable.Compile(action.Arguments[1], r.Config.Var)

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

			step++

			break
		case "COPY":
			sourceFilepath := variable.Compile(action.Arguments[0], r.Config.Var)
			destinationFilepath := variable.Compile(action.Arguments[1], r.Config.Var)

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

			step++

			break
		case "DELETE":
			args := variable.CompileArray(action.Arguments, r.Config.Var)

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

			step++

			break
		case "UPLOAD":
			sourceFiles := variable.CompileArray(action.Arguments[:len(action.Arguments)-2], r.Config.Var)
			destinationDir := variable.Compile(action.Arguments[len(action.Arguments)-1], r.Config.Var)

			if path.IsAbs(destinationDir) == false {
				if r.Config.CWD != "" {
					destinationDir = path.Join(r.Config.CWD, destinationDir)
				}
			}

			fmt.Printf("[Step %v]: UPLOAD local:%s to remote:%s\n", step, color.YellowString(strings.Join(action.Arguments, ", ")), color.GreenString(destinationDir))

			for _, filePath := range sourceFiles {

				if path.IsAbs(filePath) == false {
					filePath = path.Join(localCwd, filePath)
				}

				err := client.Upload(filePath, destinationDir)

				if err != nil {
					return err
				}
			}

			step++

			break
		case "DOWNLOAD":
			sourceFiles := variable.CompileArray(action.Arguments[:len(action.Arguments)-2], r.Config.Var)
			destinationDir := variable.Compile(action.Arguments[len(action.Arguments)-1], r.Config.Var)

			if path.IsAbs(destinationDir) == false {
				destinationDir = path.Join(localCwd, destinationDir)
			}

			fmt.Printf("[Step %v]: DOWNLOAD remote:%s to local:%s\n", step, color.YellowString(strings.Join(action.Arguments, ", ")), color.GreenString(destinationDir))

			for _, filePath := range sourceFiles {

				if path.IsAbs(filePath) == false {
					if r.Config.CWD != "" {
						filePath = path.Join(r.Config.CWD, filePath)
					}
				}

				err := client.Download(filePath, destinationDir)

				if err != nil {
					return err
				}
			}

			step++

			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Action))
		}
	}

	return nil
}
