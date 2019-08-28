package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Token struct {
	Key   string
	Value []string
}

var (
	CommentIdentifier = "#"
	ValidKeywordReg   = regexp.MustCompile("CONNECT|ENV|CD|UPLOAD|DOWNLOAD|COPY|MOVE|DELETE|RUN|CMD|BASH")
	KeywordRed        = regexp.MustCompile("[A-Z]")
	EmptyStrReg       = regexp.MustCompile("\\s")
	LineWrapReg       = regexp.MustCompile("\\\n")
)

func GenerateAST(input string) ([]Token, error) {
	currentIndex := 0

	tokens := make([]Token, 0)

	for {
		if currentIndex >= len(input)-1 {
			break
		}

		char := string(input[currentIndex])

		// if found line break. skip
		if LineWrapReg.MatchString(char) {
			currentIndex++
			continue
		}

		// if found space blank. skip
		if EmptyStrReg.MatchString(char) {
			currentIndex++
			continue
		}

		// if found comment. skip
		if char == CommentIdentifier {
		findCommentContent:
			for {
				if currentIndex > len(input)-1 {
					break findCommentContent
				}
				char = string(input[currentIndex])
				if LineWrapReg.MatchString(char) {
					break findCommentContent
				} else {
					currentIndex++
				}
			}
			continue
		}

		// if match the keyword
		if KeywordRed.MatchString(char) == true {
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

				if KeywordRed.MatchString(char) == false {
					break findKeyword
				}

				keyword += char

				currentIndex++
			}

			// valid keyword
			if ValidKeywordReg.MatchString(keyword) == false {
				return tokens, errors.New(fmt.Sprintf("invalid keyword `%s`", keyword))
			}

		skipEmptyString:
			for {
				if currentIndex > len(input)-1 {
					break
				}
				char = string(input[currentIndex])
				if EmptyStrReg.MatchString(char) {
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
				if LineWrapReg.MatchString(char) {
					break findValue
				}

				// if found comment, then ignore future content
				if char == CommentIdentifier {
					break findValue
				}

				// if in space link in value, so we think it's an other value
				if EmptyStrReg.MatchString(char) {
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
			valueStr := strings.Join(value, " ")
			valueLength := len(value)

			switch keyword {
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
