package actions

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"strings"
)

func Var(r *runner.Runner, action configuration.Action) error {
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
					return errors.New(fmt.Sprintf("run command '%s' fail.", action.Arguments))
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
}
