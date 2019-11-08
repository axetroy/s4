package command

import "github.com/axetroy/s4/core/runner"

func Default(configFile, password string, check bool) error {
	r, err := runner.NewRunner(configFile)

	if err != nil {
		return err
	}

	if password != "" {
		r.SetPassword(password)
	}

	if err := r.Run(check); err != nil {
		return err
	}

	return nil
}
