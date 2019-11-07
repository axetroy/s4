package runner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/axetroy/s4/core/configuration"
	"github.com/axetroy/s4/core/grammar"
	"github.com/axetroy/s4/core/ssh"
	"github.com/axetroy/s4/core/variable"
	"github.com/fatih/color"
)

type Runner struct {
	Config *configuration.Configuration // the configuration
	SSH    *ssh.Client                  // current ssh client
	Step   int                          // current step
	Cwd    string                       // current working dir at local
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

func (r *Runner) Run(check bool) error {
	client := ssh.NewSSH(r.Config)
	r.SSH = client

	if r.Config.Host == "" {
		return errors.New("`CONNECT` field required")
	}

	fmt.Printf("[step %v]: CONNECT %s\n", r.Step, color.GreenString(fmt.Sprintf("%s@%s:%s", r.Config.Username, r.Config.Host, r.Config.Port)))

	if check {
		return nil
	}

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

	if remoteCwd, err := client.Pwd(); err != nil {
		return err
	} else {
		r.Config.CWD = remoteCwd
	}

	for _, action := range r.Config.Actions {
		r.Step++
		switch action.Action {
		case grammar.ActionVAR:
			if err := r.actionVar(action); err != nil {
				return err
			}
			break
		case grammar.ActionCD:
			if err := r.actionCd(action); err != nil {
				return err
			}
			break
		case grammar.ActionBASH:
			if err := r.actionBash(action); err != nil {
				return err
			}
			break
		case grammar.ActionCMD:
			if err := r.actionCmd(action); err != nil {
				return err
			}
			break
		case grammar.ActionRUN:
			if err := r.actionRun(action); err != nil {
				return err
			}
			break
		case grammar.ActionMOVE:
			if err := r.actionMove(action); err != nil {
				return err
			}
			break
		case grammar.ActionCOPY:
			if err := r.actionCopy(action); err != nil {
				return err
			}
			break
		case grammar.ActionDELETE:
			if err := r.actionDelete(action); err != nil {
				return err
			}
			break
		case grammar.ActionUPLOAD:
			if err := r.actionUpload(action); err != nil {
				return err
			}
			break
		case grammar.ActionDOWNLOAD:
			if err := r.actionDownload(action); err != nil {
				return err
			}
			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Action))
		}
	}

	r.Step++

	fmt.Printf("[step %d]: %s\n", r.Step, color.GreenString("done!"))

	return nil
}

func (r *Runner) actionBash(action configuration.Action) error {
	command := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: BASH %s\n", r.Step, color.YellowString(command))

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

	command = variable.Compile(command, r.Config.Var)

	c := exec.Command(bashPath, "-c", command)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}

	if c.ProcessState.Success() == false {
		return errors.New(fmt.Sprintf("run command '%s' fail.", action.Arguments))
	}

	return nil
}

func (r *Runner) actionCd(action configuration.Action) error {
	dir := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: CD %s\n", r.Step, color.GreenString(dir))

	cwd := variable.Compile(dir, r.Config.Var)

	if path.IsAbs(cwd) {
		r.Config.CWD = cwd
	} else {
		r.Config.CWD = path.Join(r.Config.CWD, cwd)
	}

	return nil
}

func (r *Runner) actionCmd(action configuration.Action) error {
	fmt.Printf("[step %d]: CMD %s\n", r.Step, color.YellowString(fmt.Sprintf("%v", action.Arguments)))

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

	return nil
}

func (r *Runner) actionCopy(action configuration.Action) error {
	sourceFilepath := variable.Compile(action.Arguments[0], r.Config.Var)
	destinationFilepath := variable.Compile(action.Arguments[1], r.Config.Var)

	fmt.Printf("[step %d]: COPY %s to %s\n", r.Step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

	if path.IsAbs(sourceFilepath) == false {
		sourceFilepath = path.Join(r.Config.CWD, sourceFilepath)
	}

	if path.IsAbs(destinationFilepath) == false {
		destinationFilepath = path.Join(r.Config.CWD, destinationFilepath)
	}

	if err := r.SSH.Copy(sourceFilepath, destinationFilepath); err != nil {
		return err
	}
	return nil
}

func (r *Runner) actionDelete(action configuration.Action) error {
	fmt.Printf("[step %v]: DELETE %s\n", r.Step, color.YellowString(strings.Join(action.Arguments, ",")))

	args := variable.CompileArray(action.Arguments, r.Config.Var)

	var files []string

	for _, file := range args {
		if path.IsAbs(file) == false {
			file = path.Join(r.Config.CWD, file)
		}

		files = append(files, file)
	}

	if err := r.SSH.Delete(files...); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionDownload(action configuration.Action) error {
	sourceFiles := variable.CompileArray(action.Arguments[:len(action.Arguments)-1], r.Config.Var)
	destinationDir := variable.Compile(action.Arguments[len(action.Arguments)-1], r.Config.Var)

	fmt.Printf("[step %d]: DOWNLOAD %s to %s\n", r.Step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	if path.IsAbs(destinationDir) == false {
		destinationDir = path.Join(r.Cwd, destinationDir)
	}

	for _, filePath := range sourceFiles {

		if filePath == "" {
			continue
		}

		if path.IsAbs(filePath) == false {
			if r.Config.CWD != "" {
				filePath = path.Join(r.Config.CWD, filePath)
			}
		}

		err := r.SSH.Download(filePath, destinationDir)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionMove(action configuration.Action) error {
	sourceFilepath := variable.Compile(action.Arguments[0], r.Config.Var)
	destinationFilepath := variable.Compile(action.Arguments[1], r.Config.Var)

	fmt.Printf("[step %d]: MOVE %s to %s\n", r.Step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

	if path.IsAbs(sourceFilepath) == false {
		sourceFilepath = path.Join(r.Config.CWD, sourceFilepath)
	}

	if path.IsAbs(destinationFilepath) == false {
		destinationFilepath = path.Join(r.Config.CWD, destinationFilepath)
	}

	if err := r.SSH.Move(sourceFilepath, destinationFilepath); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionRun(action configuration.Action) error {
	command := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: RUN %s\n", r.Step, color.YellowString(command))

	command = variable.Compile(command, r.Config.Var)

	if err := r.SSH.Run(command); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionUpload(action configuration.Action) error {
	sourceFiles := variable.CompileArray(action.Arguments[:len(action.Arguments)-1], r.Config.Var)
	destinationDir := variable.Compile(action.Arguments[len(action.Arguments)-1], r.Config.Var)

	fmt.Printf("[step %d]: UPLOAD %s to %s\n", r.Step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	if path.IsAbs(destinationDir) == false {
		if r.Config.CWD != "" {
			destinationDir = path.Join(r.Config.CWD, destinationDir)
		}
	}

	for _, filePath := range sourceFiles {

		if filePath == "" {
			continue
		}

		if path.IsAbs(filePath) == false {
			filePath = path.Join(r.Cwd, filePath)
		}

		err := r.SSH.Upload(filePath, destinationDir)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionVar(action configuration.Action) error {
	value := strings.Join(action.Arguments, " ")
	fmt.Printf("[step %d]: VAR %s\n", r.Step, color.GreenString(value))

	if Var, err := variable.Parse(value); err != nil {
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
				remoteEnvValue, err := r.SSH.Env(Var.Value)

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
					return errors.New(fmt.Sprintf("run command '%s' fail.", Var.Value))
				}

				r.Config.Var[Var.Key] = strings.TrimSpace(stdoutBuf.String())
			} else {
				// execute command at remote
				var stdoutBuf bytes.Buffer
				var stderrBuf bytes.Buffer

				err := r.SSH.RunRaw(Var.Value, &stdoutBuf, &stderrBuf)

				if err != nil {
					return err
				}

				r.Config.Var[Var.Key] = strings.TrimSpace(stdoutBuf.String())

				fmt.Println(r.Config.Var[Var.Key])
			}
			break
		}
	}

	return nil
}
