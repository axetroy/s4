package grammar

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/axetroy/s4/src/host"
	"regexp"
	"strings"
)

type Token struct {
	Key   string
	Value []string
}

var (
	commentIdentifier = "#"
	validKeywordReg   = regexp.MustCompile("CONNECT|ENV|VAR|CD|UPLOAD|DOWNLOAD|COPY|MOVE|DELETE|RUN|CMD|BASH")
	keywordRed        = regexp.MustCompile("[A-Z]")
	emptyStrReg       = regexp.MustCompile("\\s")
	lineWrapReg       = regexp.MustCompile("\\\n")
	lineBreakChar     = "\\"
	spaceBlank        = " "
)

func isAllowLineBreakAction(actionName string) bool {
	return actionName == "RUN" || actionName == "BASH"
}

func Tokenizer(input string) ([]Token, error) {
	currentIndex := 0

	tokens := make([]Token, 0)

	for {
		if currentIndex >= len(input)-1 {
			break
		}

		char := string(input[currentIndex])

		// if found line break. skip
		if lineWrapReg.MatchString(char) {
			currentIndex++
			continue
		}

		// if found space blank. skip
		if emptyStrReg.MatchString(char) {
			currentIndex++
			continue
		}

		// if found comment. skip
		if char == commentIdentifier {
		findCommentContent:
			for {
				if currentIndex > len(input)-1 {
					break findCommentContent
				}
				char = string(input[currentIndex])
				if lineWrapReg.MatchString(char) {
					break findCommentContent
				} else {
					currentIndex++
				}
			}
			continue
		}

		// if match the keyword
		if keywordRed.MatchString(char) == true {
			var (
				keyword      = ""
				value        []string
				currentValue = ""
			)

		findKeyword:
			for {
				if currentIndex > len(input)-1 {
					break
				}
				char = string(input[currentIndex])

				if keywordRed.MatchString(char) == false {
					break findKeyword
				}

				keyword += char

				currentIndex++
			}

			// valid keyword
			if validKeywordReg.MatchString(keyword) == false {
				return tokens, errors.New(fmt.Sprintf("invalid keyword `%s`", keyword))
			}

		skipEmptyString:
			for {
				if currentIndex > len(input)-1 {
					break
				}
				char = string(input[currentIndex])
				if emptyStrReg.MatchString(char) {
					currentIndex++
				} else {
					break skipEmptyString
				}
			}

		findValue:
			for {
				if currentIndex > len(input)-1 {
					break findValue
				}

				char = string(input[currentIndex])

				// if found line wrap, eg. \n
				if lineWrapReg.MatchString(char) {

					// only allow RUN and BASH to use line break
					if isAllowLineBreakAction(keyword) {
						// find space blank forward and skip it.
						lastCharIndex := currentIndex - 1
						lastChar := ""
					findSpace:
						for {
							lastChar = string(input[lastCharIndex])

							if emptyStrReg.MatchString(lastChar) {
								lastCharIndex--
								continue
							}

							break findSpace
						}

						if lastChar == lineBreakChar {
							currentIndex++
							continue
						}
					}

					break findValue
				}

				// if found comment, then ignore future content
				if char == commentIdentifier {
					break findValue
				}

				// if in space link in value, so we think it's an other value
				if emptyStrReg.MatchString(char) {
					value = append(value, currentValue)
					currentValue = ""
					currentIndex++
					continue findValue
				}

				currentValue += char

				currentIndex++
			}

			if len(currentValue) != 0 {
				value = append(value, currentValue)
			}

			currentValue = ""

			// value must set
			if len(value) == 0 {
				return tokens, errors.New(fmt.Sprintf("`%s` require value.", keyword))
			}

			// validate token here
			valueStr := strings.Join(value, spaceBlank)
			valueLength := len(value)

			if isAllowLineBreakAction(keyword) {

				valueStr := regexp.MustCompile("\\\\\\s+").ReplaceAllString(valueStr, "")

				value = strings.Split(valueStr, spaceBlank)
			}

			switch keyword {
			case "CONNECT":
				if _, err := host.Parse(valueStr); err != nil {
					return tokens, err
				}
			case "ENV":
				if regexp.MustCompile("\\w+\\s?=\\s?\\w+").MatchString(valueStr) == false {
					return tokens, errors.New(fmt.Sprintf("`ENV` need to match `KEY = VALUE` format but got `%s`", valueStr))
				}

				newValue := []string{
					value[0],
					value[2],
				}

				value = newValue

				break

			case "CD":
				if valueLength != 1 {
					return tokens, errors.New(fmt.Sprintf("`CD` only accepts one string but got `%s`", valueStr))
				}
				break

			case "UPLOAD":
				fallthrough
			case "DOWNLOAD":
				if valueLength < 2 {
					return tokens, errors.New(fmt.Sprintf("`%s` only accepts one string but got `%s`", keyword, valueStr))
				}

				break
			case "COPY":
				fallthrough
			case "MOVE":
				if valueLength != 2 {
					return tokens, errors.New(fmt.Sprintf("`%s` only accepts two string but got `%s`", keyword, valueStr))
				}
				break
			case "DELETE":
				if valueLength < 1 {
					return tokens, errors.New(fmt.Sprintf("`%s` accepts at least one parameter but got `%s`", keyword, valueStr))
				}
				break
			case "RUN":
				fallthrough
			case "BASH":
				if valueLength < 1 {
					return tokens, errors.New(fmt.Sprintf("`%s` accepts at least one parameter but got `%s`", keyword, valueStr))
				}

				break
			case "CMD":
				var commands []string

				if err := json.Unmarshal([]byte(valueStr), &commands); err != nil {
					return tokens, errors.New(fmt.Sprintf("`%s` require JSON array format but got `%s`\n", keyword, valueStr))
				}

				value = commands

				break
			}

			tokens = append(tokens, Token{
				Key:   keyword,
				Value: value,
			})

			continue
		}

		return tokens, errors.New(fmt.Sprintf("Invalid token `%s`", char))
	}

	return tokens, nil
}
