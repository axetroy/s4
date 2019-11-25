package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/axetroy/s4/core/grammar"
	"github.com/axetroy/s4/core/host"
	"github.com/axetroy/s4/core/ssh"
	"github.com/axetroy/s4/core/variable"
	"github.com/fatih/color"
)

type Runner struct {
	ssh         *ssh.Client       // current ssh client
	totalStep   int               // total step
	currentStep int               // current step
	cwdLocal    string            // current working dir at local
	tokens      []grammar.Token   // token from parsing
	cwdRemote   string            // current remote working dir
	env         map[string]string // env for remote
	variable    map[string]string // var
}

func NewRunner(configFilepath string) (*Runner, error) {
	if f, err := os.Stat(configFilepath); err != nil {
		msg := fmt.Sprintf("Config file `%s` not found. print 's4 --help' for help.", configFilepath)
		return nil, errors.New(color.RedString(msg))
	} else {
		if f.IsDir() {
			msg := fmt.Sprintf("Config file `%s` is not a file.", configFilepath)
			return nil, errors.New(color.RedString(msg))
		}
	}

	fmt.Printf("Load the s4 file `%s`.\n", color.GreenString(configFilepath))

	content, err := ioutil.ReadFile(configFilepath)

	if err != nil {
		return nil, err
	}

	tokens, err := grammar.Tokenizer(string(content))

	if err != nil {
		return nil, err
	}

	return &Runner{
		currentStep: 1,
		totalStep:   len(tokens),
		tokens:      tokens,
		env:         map[string]string{},
		variable:    map[string]string{},
	}, nil
}

func (r *Runner) requireConnection() error {
	if r.ssh == nil {
		return errors.New("you need to connect to server first")
	} else {
		return nil
	}
}

func (r *Runner) resolveLocalPath(localPath string) string {
	if path.IsAbs(localPath) {
		return localPath
	} else {
		return path.Join(r.cwdLocal, localPath)
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
		return path.Join(r.cwdRemote, remotePath)
	}
}

func (r *Runner) resolveRemotePaths(remotePaths []string) []string {
	var paths []string

	for _, remotePath := range remotePaths {
		paths = append(paths, r.resolveRemotePath(remotePath))
	}

	return paths
}

func (r *Runner) nextStep(action string, msg string) {
	fmt.Printf("Step %d/%d: %s %s\n", r.currentStep, r.totalStep, strings.ToUpper(action), msg)
	r.currentStep++
}

func printTimeDiff(d1 time.Time, d2 time.Time) {
	timeDiffNano := d2.UnixNano() - d1.UnixNano()

	diffSecond := float64(timeDiffNano) / 1000 / 1000 / 1000

	fmt.Println(color.GreenString(fmt.Sprintf("Finish in %ss.", fmt.Sprintf("%f", diffSecond))))
}

func (r *Runner) Run() error {
	defer func() {
		if r.ssh != nil {
			_ = r.ssh.Disconnect()
		}
	}()

	d1 := time.Now()

	for _, action := range r.tokens {
		var err error
		switch action.Key {
		case grammar.ActionCONNECT:
			err = r.actionConnect(action.Node.(grammar.NodeConnect))
			break
		case grammar.ActionVAR:
			err = r.actionVar(action.Node.(grammar.NodeVar))
			break
		case grammar.ActionENV:
			err = r.actionEnv(action.Node.(grammar.NodeEnv))
			break
		case grammar.ActionCD:
			err = r.actionCd(action.Node.(grammar.NodeCd))
			break
		case grammar.ActionCMD:
			err = r.actionCmd(action.Node.(grammar.NodeCmd))
			break
		case grammar.ActionRUN:
			err = r.actionRun(action.Node.(grammar.NodeRun))
			break
		case grammar.ActionMOVE:
			err = r.actionMove(action.Node.(grammar.NodeCopy))
			break
		case grammar.ActionCOPY:
			err = r.actionCopy(action.Node.(grammar.NodeCopy))
			break
		case grammar.ActionDELETE:
			err = r.actionDelete(action.Node.(grammar.NodeDelete))
			break
		case grammar.ActionUPLOAD:
			err = r.actionUpload(action.Node.(grammar.NodeUpload))
			break
		case grammar.ActionDOWNLOAD:
			err = r.actionDownload(action.Node.(grammar.NodeUpload))
			break
		default:
			err = fmt.Errorf("invalid action `%s`", action.Key)
		}

		if err != nil {
			printTimeDiff(d1, time.Now())
			return err
		}
	}

	printTimeDiff(d1, time.Now())

	return nil
}

func (r *Runner) actionConnect(params grammar.NodeConnect) error {
	r.nextStep(grammar.ActionCONNECT, color.GreenString(fmt.Sprintf("%s@%s:%s", params.Username, params.Host, params.Port)))

	// if ssh client exist. disconnect first
	if r.ssh != nil {
		if err := r.ssh.Disconnect(); err != nil {
			return err
		}
		r.ssh = nil
	}

	var password = new(string)
	var privateKey = new([]byte)

	if params.ConnectType != nil {
		switch *params.ConnectType {
		case host.ConnectTypePassword:
			s := variable.Compile(*params.Password, r.variable)
			password = &s
			break
		case host.ConnectTypePrivateKeyFile:
			b, err := ioutil.ReadFile(*params.Password)

			if err != nil {
				return err
			}

			privateKey = &b
			break
		default:
			return fmt.Errorf("invalid connection type `%s`", *params.ConnectType)
		}
	} else {
		// ask password for remote server
		prompt := &survey.Password{
			Message: "Please type remote server's password",
		}

		if err := survey.AskOne(prompt, password); err != nil {
			return err
		}
	}

	r.ssh = ssh.NewSSH()

	if err := r.ssh.Connect(params.Host, params.Port, params.Username, password, privateKey); err != nil {
		r.ssh = nil
		return err
	}

	if cwd, err := os.Getwd(); err != nil {
		return err
	} else {
		r.cwdLocal = cwd
	}

	if remoteCwd, err := r.ssh.Pwd(); err != nil {
		return err
	} else {
		r.cwdRemote = remoteCwd
	}

	return nil
}

func (r *Runner) actionCd(params grammar.NodeCd) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	dir := params.Target

	r.nextStep(grammar.ActionCD, color.GreenString(dir))

	cwd := variable.Compile(dir, r.variable)

	r.cwdRemote = r.resolveRemotePath(cwd)

	return nil
}

func (r *Runner) actionCmd(params grammar.NodeCmd) error {
	r.nextStep(grammar.ActionCMD, color.YellowString(fmt.Sprintf("%v", params.SourceCode)))

	_, _ = os.Stderr.Write([]byte(fmt.Sprint("WARNING: `CMD` have been deprecated. it will be remove at next major version. use `RUN` instead.")))

	command := variable.Compile(params.Command, r.variable)
	args := variable.CompileArray(params.Arguments, r.variable)

	c := exec.Command(command, args...)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}

	if c.ProcessState.Success() == false {
		return fmt.Errorf("run command '%s' fail", params.SourceCode)
	}

	return nil
}

func (r *Runner) actionCopy(params grammar.NodeCopy) error {
	sourceFilepath := params.Source
	destinationFilepath := params.Destination

	r.nextStep(
		grammar.ActionCOPY,
		fmt.Sprintf(
			"%s to %s",
			color.YellowString(sourceFilepath),
			color.GreenString(destinationFilepath),
		),
	)

	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFilepath = variable.Compile(sourceFilepath, r.variable)
	destinationFilepath = variable.Compile(destinationFilepath, r.variable)

	sourceFilepath = r.resolveRemotePath(sourceFilepath)
	destinationFilepath = r.resolveRemotePath(destinationFilepath)

	if err := r.ssh.Copy(sourceFilepath, destinationFilepath); err != nil {
		return err
	}
	return nil
}

func (r *Runner) actionDelete(params grammar.NodeDelete) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	r.nextStep(grammar.ActionDELETE, color.YellowString(strings.Join(params.Targets, ", ")))

	args := variable.CompileArray(params.Targets, r.variable)

	files := r.resolveRemotePaths(args)

	if err := r.ssh.Delete(files...); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionDownload(params grammar.NodeUpload) error {
	sourceFiles := params.SourceFiles
	destinationDir := params.DestinationDir

	r.nextStep(
		grammar.ActionDOWNLOAD,
		fmt.Sprintf(
			"%s to %s",
			color.YellowString(strings.Join(sourceFiles, ", ")),
			color.GreenString(destinationDir),
		),
	)

	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFiles = variable.CompileArray(sourceFiles, r.variable)
	destinationDir = variable.Compile(destinationDir, r.variable)

	sourceFiles = r.resolveRemotePaths(sourceFiles)
	destinationDir = r.resolveLocalPath(destinationDir)

	for _, filePath := range sourceFiles {
		if err := r.ssh.Download(filePath, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionMove(params grammar.NodeCopy) error {
	sourceFilepath := params.Source
	destinationFilepath := params.Destination

	r.nextStep(
		grammar.ActionMOVE,
		fmt.Sprintf(
			"%s to %s",
			color.YellowString(sourceFilepath),
			color.GreenString(destinationFilepath),
		),
	)

	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFilepath = variable.Compile(sourceFilepath, r.variable)
	destinationFilepath = variable.Compile(destinationFilepath, r.variable)

	sourceFilepath = r.resolveRemotePath(sourceFilepath)
	destinationFilepath = r.resolveRemotePath(destinationFilepath)

	if err := r.ssh.Move(sourceFilepath, destinationFilepath); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionRun(params grammar.NodeRun) error {
	r.nextStep(grammar.ActionRUN, color.YellowString(params.SourceCode))

	isPipeCommand := len(params.Commands) > 1
	var lastCommandStdout bytes.Buffer

	for _, cmd := range params.Commands {
		if cmd.RunInLocal {
			command := variable.Compile(cmd.Command[0], r.variable)
			args := variable.CompileArray(cmd.Command[1:], r.variable)

			c := exec.Command(command, args...)

			c.Stdin = bytes.NewReader(lastCommandStdout.Bytes())
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				return err
			}

			if c.ProcessState.Success() == false {
				return fmt.Errorf("run command '%v' fail", params.SourceCode)
			}
		} else {
			if err := r.requireConnection(); err != nil {
				return err
			}

			command := variable.Compile(cmd.SourceCode, r.variable)

			if stdout, _, err := r.ssh.Run(command, ssh.Options{CWD: r.cwdRemote, Env: r.env}); err != nil {
				return err
			} else {
				if isPipeCommand {
					lastCommandStdout = stdout
				}
			}
		}
	}

	return nil
}

func (r *Runner) actionUpload(params grammar.NodeUpload) error {
	sourceFiles := params.SourceFiles
	destinationDir := params.DestinationDir

	r.nextStep(
		grammar.ActionUPLOAD,
		fmt.Sprintf(
			"%s to %s",
			color.YellowString(strings.Join(sourceFiles, ", ")),
			color.GreenString(destinationDir),
		),
	)

	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFiles = variable.CompileArray(sourceFiles, r.variable)
	destinationDir = variable.Compile(destinationDir, r.variable)

	sourceFiles = r.resolveLocalPaths(sourceFiles)
	destinationDir = r.resolveRemotePath(destinationDir)

	for _, filePath := range sourceFiles {
		if err := r.ssh.Upload(filePath, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionEnv(params grammar.NodeEnv) error {
	r.nextStep(grammar.ActionENV, color.GreenString(params.SourceCode))
	r.env[params.Key] = variable.Compile(params.Value, r.variable)
	return nil
}

func (r *Runner) actionVar(params grammar.NodeVar) error {
	r.nextStep(grammar.ActionVAR, color.GreenString(params.SourceCode))

	if params.Literal != nil {
		r.variable[params.Key] = params.Literal.Value
	} else if params.Env != nil {
		if params.Env.Local {
			r.variable[params.Key] = os.Getenv(variable.Compile(params.Env.Key, r.variable))
		} else {
			if err := r.requireConnection(); err != nil {
				return err
			}
			if remoteEnvValue, err := r.ssh.Env(variable.Compile(params.Env.Key, r.variable), ssh.Options{Env: r.env}); err != nil {
				return err
			} else {
				r.variable[params.Key] = remoteEnvValue
			}
		}
	} else if params.Command != nil {
		if params.Command.Local {
			commandArr := variable.CompileArray(params.Command.Command, r.variable)

			command := commandArr[0]
			args := commandArr[1:]

			c := exec.Command(command, args...)

			var stdoutBuf bytes.Buffer
			var stderrBuf bytes.Buffer

			c.Stdout = &stdoutBuf
			c.Stderr = &stderrBuf

			if err := c.Run(); err != nil {
				return err
			}

			if c.ProcessState.Success() == false {
				return fmt.Errorf("run command '%s' fail", params.Command.Command)
			}

			r.variable[params.Key] = strings.TrimSpace(stdoutBuf.String())
		} else {
			if err := r.requireConnection(); err != nil {
				return err
			}

			stdout, _, err := r.ssh.Run(strings.Join(params.Command.Command, " "), ssh.Options{
				CWD: r.cwdRemote,
				Env: r.env,
			})

			if err != nil {
				return err
			}

			output := stdout.String()

			r.variable[params.Key] = strings.TrimSpace(output)
		}
	}

	return nil
}
