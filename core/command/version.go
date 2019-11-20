package command

import "os"

// Version Print version to stdout
func Version(version string) error {
	_, err := os.Stdout.Write([]byte(version))

	return err
}
