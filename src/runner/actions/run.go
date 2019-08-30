package actions

import (
	"fmt"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"strings"
)

func Run(r *runner.Runner, action configuration.Action) error {
	command := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: RUN %s\n", r.Step, color.YellowString(command))

	command = variable.Compile(command, r.Config.Var)

	if err := r.SSH.Run(command); err != nil {
		return err
	}

	return nil
}
