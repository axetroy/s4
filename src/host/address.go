package host

import (
	"fmt"
	"github.com/kataras/iris/core/errors"
	"regexp"
)

type Address struct {
	Host     string
	Port     string
	Username string
}

var (
	addressReg = regexp.MustCompile("^([\\w-]+)@([\\w\\.-]+):(\\d+)$")
)

func Parse(address string) (Address, error) {
	addr := Address{}

	matchers := addressReg.FindAllStringSubmatch(address, -1)

	if len(matchers) == 0 {
		return addr, errors.New(fmt.Sprintf("Invalid address `%s`", address))
	}

	matcher := matchers[0]
	username := matcher[1]
	host := matcher[2]
	port := matcher[3]

	addr.Host = host
	addr.Port = port
	addr.Username = username

	return addr, nil
}
