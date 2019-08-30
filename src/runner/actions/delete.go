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

func Delete(r *runner.Runner, action configuration.Action) error {
	fmt.Printf("[step %v]: DELETE %s\n", r.Step, color.YellowString(strings.Join(action.Arguments, ",")))

	args := variable.CompileArray(action.Arguments, r.Config.Var)

	var files []string

	for _, file := range args {
		if path.IsAbs(file) == false {
			file = path.Join(r.Config.CWD, file)
		}

		files = append(files, file)
	}

	if err := r.SSH.Delete(files...); err != nil {
		return err
	}

	return nil
}
