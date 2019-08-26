package parser

import (
	"errors"
	"fmt"
	"github.com/axetroy/go-fs"
	"github.com/fatih/color"
	"strconv"
	"strings"
)

type Action struct {
	Action    string
	Arguments string
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	CWD      string
	Actions  []Action
}

func Parse(configFile string) (c Config, err error) {
	var content []byte
	content, err = fs.ReadFile(configFile)

	raw := string(content[:])

	lines := strings.Split(raw, "\n")

	for _, line := range lines {
		s := strings.Trim(line, "")
		if s == "" {
			continue
		}

		arr := strings.Split(s, " ")

		keyword := arr[0]
		value := strings.Join(arr[1:], " ")

		switch keyword {
		case "HOST":
			c.Host = value
			break
		case "PORT":
			if port, e := strconv.Atoi(value); e != nil {
				err = e
				return
			} else {
				c.Port = port
			}
			break
		case "USERNAME":
			c.Username = value
			break
		case "CWD":
			c.CWD = value
			break
		case "COPY":
			c.Actions = append(c.Actions, Action{
				Action:    keyword,
				Arguments: value,
			})
			break
		case "RUN":
			c.Actions = append(c.Actions, Action{
				Action:    keyword,
				Arguments: value,
			})
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid keyword `%s`", color.RedString(keyword)))
			return
		}

	}

	return
}
