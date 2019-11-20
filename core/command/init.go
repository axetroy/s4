package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const (
	defaultTemplate = `# This is an example of using s4.
# For more detail: https://github.com/axetroy/s4
CONNECT root@192.168.0.0.1:22

ENV	FOO = bar

VAR	name = s4

CD /root

RUN echo "project name: {{name}}"

RUN echo "current foo environmental variable: $FOO"

RUN npm run lint \
	&& npm run test \
	&& npm run build
`
)

/**
check a path is exist or not
*/
func pathExists(path string) (isExist bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

/*
write a file
*/
func writeFileStr(filepath string, data string) error {
	return ioutil.WriteFile(filepath, []byte(data), os.ModePerm)
}

// Init a s4 file
func Init() error {
	var (
		cwd string
		err error
	)

	if cwd, err = os.Getwd(); err != nil {
		return err
	}

	filepath := path.Join(cwd, ".s4")

	if pathExists(filepath) {
		fmt.Printf("s4 file `%s` already exist.\n", filepath)
		return nil
	}

	if err = writeFileStr(filepath, defaultTemplate); err != nil {
		return err
	}

	fmt.Println("s4 file have been create.")

	return nil
}
