package command

import "github.com/axetroy/s4/src/runner"

func Detault(configFile, password string) error {
	r, err := runner.NewRunner(configFile)

	if err != nil {
		return err
	}

	if password != "" {
		r.Config.Password = password
	}

	if err := r.Run(); err != nil {
		return err
	}

	return nil
}
