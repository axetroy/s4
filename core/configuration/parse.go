package configuration

import (
	"github.com/axetroy/s4/core/grammar"
	"github.com/axetroy/s4/core/host"
	"io/ioutil"
	"strings"
)

type Action struct {
	Action    string
	Arguments []string
}

type Configuration struct {
	Host     string
	Port     string
	CWD      string
	Env      map[string]string
	Var      map[string]string
	Username string
	Password string
	Actions  []Action
}

func ParseFile(configFilepath string) (*Configuration, error) {
	var content []byte
	content, err := ioutil.ReadFile(configFilepath)

	if err != nil {
		return nil, err
	}

	return Parse(content)
}

func Parse(content []byte) (c *Configuration, err error) {
	c = &Configuration{}

	c.Env = map[string]string{}
	c.Var = map[string]string{}

	tokens, err := grammar.Tokenizer(string(content))

	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		value := strings.Join(token.Value, " ")
		switch token.Key {
		case grammar.ActionCONNECT:
			addr, err := host.Parse(value)

			if err != nil {
				return nil, err
			}

			c.Host = addr.Host
			c.Port = addr.Port
			c.Username = addr.Username
			break
		case grammar.ActionENV:
			envKey := token.Value[0]
			envValue := token.Value[1]

			c.Env[envKey] = envValue
			break
		case grammar.ActionVAR:
			fallthrough
		case grammar.ActionCD:
			fallthrough
		case grammar.ActionCOPY:
			fallthrough
		case grammar.ActionMOVE:
			fallthrough
		case grammar.ActionDELETE:
			fallthrough
		case grammar.ActionUPLOAD:
			fallthrough
		case grammar.ActionDOWNLOAD:
			c.Actions = append(c.Actions, Action{
				Action:    token.Key,
				Arguments: token.Value,
			})
			break
		case grammar.ActionBASH:
			fallthrough
		case grammar.ActionCMD:
			fallthrough
		case grammar.ActionRUN:
			c.Actions = append(c.Actions, Action{
				Action:    token.Key,
				Arguments: token.Value,
			})
			break
		}

	}

	return
}
