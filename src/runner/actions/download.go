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

func Download(r *runner.Runner, action configuration.Action) error {
	sourceFiles := variable.CompileArray(action.Arguments[:len(action.Arguments)-2], r.Config.Var)
	destinationDir := variable.Compile(action.Arguments[len(action.Arguments)-1], r.Config.Var)

	fmt.Printf("[step %d]: DOWNLOAD remote:%s to local:%s\n", r.Step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	if path.IsAbs(destinationDir) == false {
		destinationDir = path.Join(r.Cwd, destinationDir)
	}

	for _, filePath := range sourceFiles {

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
