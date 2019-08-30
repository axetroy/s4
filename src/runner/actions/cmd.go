package actions

import (
	"errors"
	"fmt"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"os"
	"os/exec"
)

func Cmd(r *runner.Runner, action configuration.Action) error {
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
