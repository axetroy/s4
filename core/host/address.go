package host

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ConnectTypePassword       = "PASSWORD"
	ConnectTypePrivateKeyFile = "FILE"
	ConnectTypes              = []string{
		ConnectTypePassword,
		ConnectTypePrivateKeyFile,
	}
)

type Address struct {
	Host        string
	Port        string
	Username    string
	ConnectType *string
	Password    *string
}

var (
	addressReg = regexp.MustCompile(fmt.Sprintf("^([\\w-\\.]+)@([\\w\\.-]+):(\\d+)\\s*(WITH\\s+(%s)\\s+(.*))?$", strings.Join(ConnectTypes, "|")))
)

func Parse(address string) (Address, error) {
	addr := Address{}

	matchers := addressReg.FindAllStringSubmatch(address, -1)

	if len(matchers) == 0 {
		return addr, errors.New(fmt.Sprintf("address format should follow `<username>@<host>:<port> [WITH [%s] [VALUE]]` but got `%s`", strings.Join(ConnectTypes, "|"), address))
	}

	matcher := matchers[0]

	username := matcher[1]
	host := matcher[2]
	port := matcher[3]
	connectType := matcher[5]
	password := matcher[6]

	addr.Host = host
	addr.Port = port
	addr.Username = username

	if strings.TrimSpace(connectType) != "" {
		addr.ConnectType = &connectType
	}

	if strings.TrimSpace(password) != "" {
		addr.Password = &password
	}

	return addr, nil
}
