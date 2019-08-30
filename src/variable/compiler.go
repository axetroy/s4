package variable

import (
	"fmt"
	"regexp"
)

var (
	expressionReg = regexp.MustCompile("\\{\\{\\s*(\\w+)\\s*\\}\\}")
)

func Compile(template string, varMap map[string]string) string {
	matchers := expressionReg.FindAllStringSubmatch(template, -1)

	for _, matcher := range matchers {
		key := matcher[1]
		value := varMap[key]

		replaceReg := regexp.MustCompile(fmt.Sprintf("\\{\\{\\s*(%s)\\s*\\}\\}", key))

		template = replaceReg.ReplaceAllString(template, value)
	}

	return template
}

func CompileArray(templates []string, varMap map[string]string) []string {
	var result []string

	for _, template := range templates {
		result = append(result, Compile(template, varMap))
	}

	return result
}
