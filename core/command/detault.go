package command

import "github.com/axetroy/s4/core/runner"

func Default(configFile string) error {
	r, err := runner.NewRunner(configFile)

	if err != nil {
		return err
	}

	if err := r.Run(); err != nil {
		return err
	}

	return nil
}
