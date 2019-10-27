package command

import "os"

func Version(version string) error {
	_, err := os.Stdout.Write([]byte(version))

	return err
}
