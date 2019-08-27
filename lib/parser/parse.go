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
	CWD      string
	Env      map[string]string
	Username string
	Password string
	Actions  []Action
}

func RemoveComment(value string) string {
	hashIndex := strings.Index(value, "#")

	if hashIndex < 0 {
		return value
	}

	result := strings.Join(strings.Split(value, "")[:hashIndex], "")

	return strings.Trim(result, " ")
}

func ParseFile(s4File string) (*Config, error) {
	var content []byte
	content, err := fs.ReadFile(s4File)

	if err != nil {
		return nil, err
	}

	c, err := Parse(content)

	if err != nil {
		return nil, err
	}

	if err := Check(c); err != nil {
		return nil, err
	}

	return &c, nil
}

func Parse(content []byte) (c Config, err error) {
	raw := string(content[:])
	lines := strings.Split(raw, "\n")

	c.Env = map[string]string{}

	for _, line := range lines {
		s := strings.Trim(line, "")
		if s == "" {
			continue
		}

		// comment line
		if strings.Index(s, "#") == 0 {
			continue
		}

		arr := strings.Split(s, " ")

		keyword := arr[0]
		value := strings.Join(arr[1:], " ")

		switch keyword {
		case "HOST":
			c.Host = RemoveComment(value)
			break
		case "PORT":
			if port, e := strconv.Atoi(RemoveComment(value)); e != nil {
				err = e
				return
			} else {
				c.Port = port
			}
			break
		case "USERNAME":
			c.Username = RemoveComment(value)
			break
		case "CWD":
			if c.CWD == "" {
				c.CWD = RemoveComment(value)
			}
			fallthrough
		case "COPY":
			fallthrough
		case "MOVE":
			fallthrough
		case "DELETE":
			fallthrough
		case "UPLOAD":
			fallthrough
		case "DOWNLOAD":
			c.Actions = append(c.Actions, Action{
				Action:    keyword,
				Arguments: RemoveComment(value),
			})
			break
		case "ENV":
			keyValuePair := strings.Split(RemoveComment(strings.Trim(value, " ")), "=")

			if len(keyValuePair) < 2 {
				err = errors.New(fmt.Sprintf("Invalid ENV '%s'", value))
				return
			}

			envKey := strings.Trim(keyValuePair[0], " ")
			envValue := strings.Trim(strings.Join(keyValuePair[1:], "="), " ")

			c.Env[envKey] = envValue
			break
		case "BASH":
			fallthrough
		case "CMD":
			fallthrough
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

	// check config is valid

	return
}

// check config file is valid of not
func Check(c Config) error {
	if c.Host == "" {
		return errors.New(fmt.Sprintf("Invalid 'host' %s", c.Host))
	}

	if c.Port == 0 {
		return errors.New(fmt.Sprintf("Invalid 'port' %d", c.Port))
	}

	return nil
}
