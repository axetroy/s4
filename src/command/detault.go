package command

import "github.com/axetroy/s4/src/runner"

func Default(configFile, password string, check bool) error {
	r, err := runner.NewRunner(configFile)

	if err != nil {
		return err
	}

	if password != "" {
		r.Config.Password = password
	}

	if err := r.Run(check); err != nil {
		return err
	}

	return nil
}
