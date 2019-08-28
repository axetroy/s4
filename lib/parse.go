package lib

import (
	"io/ioutil"
	"strings"
)

type Action struct {
	Action    string
	Arguments []string
}

type Config struct {
	Host     string
	Port     string
	CWD      string
	Env      map[string]string
	Username string
	Password string
	Actions  []Action
}

func ParseFile(configFilepath string) (*Config, error) {
	var content []byte
	content, err := ioutil.ReadFile(configFilepath)

	if err != nil {
		return nil, err
	}

	return Parse(content)
}

func Parse(content []byte) (c *Config, err error) {
	c = &Config{}

	c.Env = map[string]string{}

	tokens, err := GenerateAST(string(content))

	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		value := strings.Join(token.Value, " ")
		switch token.Key {
		case "CONNECT":
			addr, err := ParseAddress(value)

			if err != nil {
				return nil, err
			}

			c.Host = addr.Host
			c.Port = addr.Port
			c.Username = addr.Username
			break
		case "ENV":
			envKey := token.Value[0]
			envValue := token.Value[1]

			c.Env[envKey] = envValue
			break
		case "CD":
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
				Action:    token.Key,
				Arguments: token.Value,
			})
			break
		case "BASH":
			fallthrough
		case "CMD":
			fallthrough
		case "RUN":
			c.Actions = append(c.Actions, Action{
				Action:    token.Key,
				Arguments: token.Value,
			})
			break
		}

	}

	return
}
