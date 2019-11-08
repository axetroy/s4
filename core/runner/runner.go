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
	config *configuration.Configuration // the configuration
	ssh    *ssh.Client                  // current ssh client
	step   int                          // current step
	cwd    string                       // current working dir at local
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
		config: config,
		step:   1,
	}, nil
}

func (r *Runner) resolveLocalPath(localPath string) string {
	if path.IsAbs(localPath) {
		return localPath
	} else {
		return path.Join(r.cwd, localPath)
	}
}

func (r *Runner) resolveLocalPaths(localPaths []string) []string {
	var paths []string

	for _, remotePath := range localPaths {
		paths = append(paths, r.resolveLocalPath(remotePath))
	}

	return paths
}

func (r *Runner) resolveRemotePath(remotePath string) string {
	if path.IsAbs(remotePath) {
		return remotePath
	} else {
		return path.Join(r.config.CWD, remotePath)
	}
}

func (r *Runner) resolveRemotePaths(remotePaths []string) []string {
	var paths []string

	for _, remotePath := range remotePaths {
		paths = append(paths, r.resolveRemotePath(remotePath))
	}

	return paths
}

func (r *Runner) Run(check bool) error {
	client := ssh.NewSSH()
	r.ssh = client

	if r.config.Host == "" {
		return errors.New("`CONNECT` field required")
	}

	fmt.Printf("[step %v]: CONNECT %s\n", r.step, color.GreenString(fmt.Sprintf("%s@%s:%s", r.config.Username, r.config.Host, r.config.Port)))

	if check {
		return nil
	}

	if r.config.Password == "" {
		// ask password for remote server
		password := ""
		prompt := &survey.Password{
			Message: "Please type remote server's password",
		}

		if err := survey.AskOne(prompt, &password); err != nil {
			return err
		}

		r.config.Password = password
	}

	if err := client.Connect(r.config.Host, r.config.Port, r.config.Username, r.config.Password); err != nil {
		return err
	}

	defer client.Disconnect()

	if cwd, err := os.Getwd(); err != nil {
		return err
	} else {
		r.cwd = cwd
	}

	if remoteCwd, err := client.Pwd(); err != nil {
		return err
	} else {
		r.config.CWD = remoteCwd
	}

	for _, action := range r.config.Actions {
		r.step++
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

	r.step++

	fmt.Printf("[step %d]: %s\n", r.step, color.GreenString("done!"))

	return nil
}

func (r *Runner) actionBash(action configuration.Action) error {
	command := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: BASH %s\n", r.step, color.YellowString(command))

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

	command = variable.Compile(command, r.config.Var)

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

	fmt.Printf("[step %d]: CD %s\n", r.step, color.GreenString(dir))

	cwd := variable.Compile(dir, r.config.Var)

	r.config.CWD = r.resolveRemotePath(cwd)

	return nil
}

func (r *Runner) actionCmd(action configuration.Action) error {
	fmt.Printf("[step %d]: CMD %s\n", r.step, color.YellowString(fmt.Sprintf("%v", action.Arguments)))

	command := variable.Compile(action.Arguments[0], r.config.Var)
	args := variable.CompileArray(action.Arguments[1:], r.config.Var)

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
	sourceFilepath := action.Arguments[0]
	destinationFilepath := action.Arguments[1]

	fmt.Printf("[step %d]: COPY %s to %s\n", r.step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

	sourceFilepath = variable.Compile(sourceFilepath, r.config.Var)
	destinationFilepath = variable.Compile(destinationFilepath, r.config.Var)

	sourceFilepath = r.resolveRemotePath(sourceFilepath)
	destinationFilepath = r.resolveRemotePath(destinationFilepath)

	if err := r.ssh.Copy(sourceFilepath, destinationFilepath); err != nil {
		return err
	}
	return nil
}

func (r *Runner) actionDelete(action configuration.Action) error {
	fmt.Printf("[step %v]: DELETE %s\n", r.step, color.YellowString(strings.Join(action.Arguments, ",")))

	args := variable.CompileArray(action.Arguments, r.config.Var)

	files := r.resolveRemotePaths(args)

	if err := r.ssh.Delete(files...); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionDownload(action configuration.Action) error {
	sourceFiles := action.Arguments[:len(action.Arguments)-1]
	destinationDir := action.Arguments[len(action.Arguments)-1]

	fmt.Printf("[step %d]: DOWNLOAD %s to %s\n", r.step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	sourceFiles = variable.CompileArray(sourceFiles, r.config.Var)
	destinationDir = variable.Compile(destinationDir, r.config.Var)

	sourceFiles = r.resolveRemotePaths(sourceFiles)
	destinationDir = r.resolveLocalPath(destinationDir)

	for _, filePath := range sourceFiles {
		if err := r.ssh.Download(filePath, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionMove(action configuration.Action) error {
	sourceFilepath := action.Arguments[0]
	destinationFilepath := action.Arguments[1]

	fmt.Printf("[step %d]: MOVE %s to %s\n", r.step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

	sourceFilepath = variable.Compile(sourceFilepath, r.config.Var)
	destinationFilepath = variable.Compile(destinationFilepath, r.config.Var)

	sourceFilepath = r.resolveRemotePath(sourceFilepath)
	destinationFilepath = r.resolveRemotePath(destinationFilepath)

	if err := r.ssh.Move(sourceFilepath, destinationFilepath); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionRun(action configuration.Action) error {
	command := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: RUN %s\n", r.step, color.YellowString(command))

	command = variable.Compile(command, r.config.Var)

	if err := r.ssh.Run(command, ssh.Options{
		CWD: r.config.CWD,
		Env: r.config.Env,
	}); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionUpload(action configuration.Action) error {
	sourceFiles := action.Arguments[:len(action.Arguments)-1]
	destinationDir := action.Arguments[len(action.Arguments)-1]

	fmt.Printf("[step %d]: UPLOAD %s to %s\n", r.step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	sourceFiles = variable.CompileArray(sourceFiles, r.config.Var)
	destinationDir = variable.Compile(destinationDir, r.config.Var)

	sourceFiles = r.resolveLocalPaths(sourceFiles)
	destinationDir = r.resolveRemotePath(destinationDir)

	for _, filePath := range sourceFiles {
		if err := r.ssh.Upload(filePath, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionVar(action configuration.Action) error {
	value := strings.Join(action.Arguments, " ")
	fmt.Printf("[step %d]: VAR %s\n", r.step, color.GreenString(value))

	if Var, err := variable.Parse(value); err != nil {
		return err
	} else {
		switch Var.Type {
		case variable.TypeLiteral:
			r.config.Var[Var.Key] = Var.Value
			break
		case variable.TypeEnv:
			if Var.Remote == false {
				// get local env
				r.config.Var[Var.Key] = os.Getenv(Var.Value)
			} else {
				// get remote env
				remoteEnvValue, err := r.ssh.Env(Var.Value, ssh.Options{Env: r.config.Env})

				if err != nil {
					return err
				}

				r.config.Var[Var.Key] = remoteEnvValue
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

				r.config.Var[Var.Key] = strings.TrimSpace(stdoutBuf.String())
			} else {
				// execute command at remote
				var stdoutBuf bytes.Buffer
				var stderrBuf bytes.Buffer

				err := r.ssh.RunRaw(Var.Value, ssh.Options{
					CWD: r.config.CWD,
					Env: r.config.Env,
				}, &stdoutBuf, &stderrBuf)

				if err != nil {
					return err
				}

				r.config.Var[Var.Key] = strings.TrimSpace(stdoutBuf.String())

				fmt.Println(r.config.Var[Var.Key])
			}
			break
		}
	}

	return nil
}
