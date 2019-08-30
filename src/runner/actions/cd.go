package actions

import (
	"fmt"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"strings"
)

func Cd(r *runner.Runner, action configuration.Action) error {
	dir := strings.Join(action.Arguments, " ")

	fmt.Printf("[step %d]: CD %s\n", r.Step, color.GreenString(dir))

	r.Config.CWD = variable.Compile(dir, r.Config.Var)

	return nil
}
