package actions

import (
	"fmt"
	"github.com/axetroy/s4/src/configuration"
	"github.com/axetroy/s4/src/runner"
	"github.com/axetroy/s4/src/variable"
	"github.com/fatih/color"
	"path"
	"strings"
)

func Upload(r *runner.Runner, action configuration.Action) error {
	sourceFiles := variable.CompileArray(action.Arguments[:len(action.Arguments)-2], r.Config.Var)
	destinationDir := variable.Compile(action.Arguments[len(action.Arguments)-1], r.Config.Var)

	fmt.Printf("[step %d]: UPLOAD local:%s to remote:%s\n", r.Step, color.YellowString(strings.Join(action.Arguments, ", ")), color.GreenString(destinationDir))

	if path.IsAbs(destinationDir) == false {
		if r.Config.CWD != "" {
			destinationDir = path.Join(r.Config.CWD, destinationDir)
		}
	}

	for _, filePath := range sourceFiles {

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
