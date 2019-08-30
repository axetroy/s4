package variable

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Type int

const (
	TypeLiteral Type = 0
	TypeEnv          = 1
	TypeCommand      = 2
)

type Variable struct {
	Key    string
	Value  string
	Type   Type
	Remote bool // If running with a command or get environmental variable, is it running remotely?
}

var (
	formatReg = regexp.MustCompile("^(\\w+)\\s*(<?=)\\s*(.*)")
)

func Parse(input string) (Variable, error) {

	v := Variable{}

	if formatReg.MatchString(input) == false {
		return v, errors.New(fmt.Sprintf("Invalid variable format `%s`", input))
	}

	m := formatReg.FindStringSubmatch(input)

	key := m[1]
	action := m[2]
	value := strings.TrimSpace(m[3])

	v.Key = key

	switch action {
	case "=":
		// set env
		envReg := regexp.MustCompile("^\\$([A-Z]+):([a-z]+)\\s*$")

		if envReg.MatchString(value) {
			matchers := envReg.FindStringSubmatch(value)

			tag := matchers[2]

			if tag == "local" {
				v.Remote = false
			} else if tag == "remote" {
				v.Remote = true
			} else {
				return v, errors.New(fmt.Sprintf("Invalid env tag `%s`", tag))
			}

			v.Type = TypeEnv
			v.Value = matchers[1]
		} else {
			v.Type = TypeLiteral
			v.Value = value
		}
		return v, nil
	case "<=":
		v.Type = TypeCommand
		// if command defined as JSON array. eg ["npm"]. this should run in local
		if strings.Index(value, "[") == 0 {
			var commands []string

			if err := json.Unmarshal([]byte(value), &commands); err != nil {
				return v, errors.New(fmt.Sprintf("Invalid JSON array format `%s`", value))
			}

			v.Value = strings.Join(commands, " ")
			v.Remote = false
		} else {
			v.Value = value
			v.Remote = true
		}
		return v, nil
	default:
		return v, errors.New(fmt.Sprintf("invalid format for variable `%s`", input))
	}
}
