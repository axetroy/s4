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
	"strings"
)

func Bash(r *runner.Runner, action configuration.Action) error {
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
