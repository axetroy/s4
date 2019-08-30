package actions

import (
	"fmt"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"path"
)

func Move(r *runner.Runner, action configuration.Action) error {
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
