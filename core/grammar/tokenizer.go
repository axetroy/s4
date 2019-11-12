package grammar

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/axetroy/s4/core/variable"
	"regexp"
	"strings"

	"github.com/axetroy/s4/core/host"
)

type Token struct {
	Key  string
	Node interface{}
}

type NodeUpload struct {
	SourceFiles    []string
	DestinationDir string
	SourceCode     string
}

type NodeConnect struct {
	Host       string
	Port       string
	Username   string
	SourceCode string
}

type NodeEnv struct {
	Key        string
	Value      string
	SourceCode string
}

type NodeVar struct {
	Key        string
	Literal    *NodeVarLiteral
	Env        *NodeVarEnv
	Command    *NodeVarCommand
	SourceCode string
}

type NodeVarLiteral struct {
	Value string
}

type NodeVarEnv struct {
	Local bool
	Key   string
}

type NodeVarCommand struct {
	Local   bool
	Command []string
}

type NodeCopy struct {
	Source      string
	Destination string
	SourceCode  string
}

type NodeCmd struct {
	Command    string
	Arguments  []string
	SourceCode string
}

type NodeBash struct {
	Command    string
	SourceCode string
}

type NodeDelete struct {
	Targets    []string
	SourceCode string
}

type NodeCd struct {
	Target     string
	SourceCode string
}

const (
	ActionCONNECT  = "CONNECT"
	ActionENV      = "ENV"
	ActionVAR      = "VAR"
	ActionCD       = "CD"
	ActionUPLOAD   = "UPLOAD"
	ActionDOWNLOAD = "DOWNLOAD"
	ActionCOPY     = "COPY"
	ActionMOVE     = "MOVE"
	ActionDELETE   = "DELETE"
	ActionRUN      = "RUN"
	ActionCMD      = "CMD"
	ActionBASH     = "BASH"
)

var (
	Actions = []string{
		ActionCONNECT,
		ActionENV,
		ActionVAR,
		ActionCD,
		ActionUPLOAD,
		ActionDOWNLOAD,
		ActionCOPY,
		ActionMOVE,
		ActionDELETE,
		ActionRUN,
		ActionRUN,
		ActionCMD,
		ActionBASH,
	}
	commentIdentifier = "#"
	validKeywordReg   = regexp.MustCompile(strings.Join(Actions, "|"))
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
			case ActionCONNECT:
				if addr, err := host.Parse(valueStr); err != nil {
					return tokens, err
				} else {
					tokens = append(tokens, Token{
						Key: keyword,
						Node: NodeConnect{
							Host:       addr.Host,
							Port:       addr.Port,
							Username:   addr.Username,
							SourceCode: valueStr,
						},
					})
				}
				break
			case ActionENV:
				if regexp.MustCompile("\\w+\\s?=\\s?\\w+").MatchString(valueStr) == false {
					return tokens, errors.New(fmt.Sprintf("`ENV` need to match `KEY = VALUE` format but got `%s`", valueStr))
				}

				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeEnv{
						Key:        value[0],
						Value:      value[2],
						SourceCode: valueStr,
					},
				})

				break
			case ActionCD:
				if valueLength != 1 {
					return tokens, errors.New(fmt.Sprintf("`CD` only accepts one string but got `%s`", valueStr))
				}
				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeCd{
						Target:     valueStr,
						SourceCode: valueStr,
					},
				})
				break
			case ActionUPLOAD:
				fallthrough
			case ActionDOWNLOAD:
				if valueLength < 2 {
					return tokens, errors.New(fmt.Sprintf("`%s` only accepts one string but got `%s`", keyword, valueStr))
				}

				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeUpload{
						SourceFiles:    value[:len(value)-1],
						DestinationDir: value[len(value)-1],
						SourceCode:     valueStr,
					},
				})

				break
			case ActionCOPY:
				fallthrough
			case ActionMOVE:
				if valueLength != 2 {
					return tokens, errors.New(fmt.Sprintf("`%s` only accepts two string but got `%s`", keyword, valueStr))
				}
				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeCopy{
						Source:      value[0],
						Destination: value[1],
						SourceCode:  valueStr,
					},
				})
				break
			case ActionDELETE:
				if valueLength < 1 {
					return tokens, errors.New(fmt.Sprintf("`%s` accepts at least one parameter but got `%s`", keyword, valueStr))
				}
				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeDelete{
						Targets:    value,
						SourceCode: valueStr,
					},
				})
				break
			case ActionRUN:
				fallthrough
			case ActionBASH:
				if valueLength < 1 {
					return tokens, errors.New(fmt.Sprintf("`%s` accepts at least one parameter but got `%s`", keyword, valueStr))
				}
				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeBash{
						Command:    valueStr,
						SourceCode: valueStr,
					},
				})

				break
			case ActionCMD:
				var commands []string

				if err := json.Unmarshal([]byte(valueStr), &commands); err != nil {
					return tokens, errors.New(fmt.Sprintf("`%s` require JSON array format but got `%s`\n", keyword, valueStr))
				}

				value = commands

				tokens = append(tokens, Token{
					Key: keyword,
					Node: NodeCmd{
						Command:    commands[0],
						Arguments:  commands[1:],
						SourceCode: valueStr,
					},
				})

				break

			case ActionVAR:
				varNode := NodeVar{}

				if Var, err := variable.Parse(valueStr); err != nil {
					return nil, err
				} else {
					varNode.Key = Var.Key
					varNode.SourceCode = valueStr

					switch Var.Type {
					case variable.TypeLiteral:
						varNode.Literal = &NodeVarLiteral{
							Value: Var.Value,
						}
						break
					case variable.TypeEnv:
						varNode.Env = &NodeVarEnv{
							Local: !Var.Remote,
							Key:   Var.Value,
						}
						break
					case variable.TypeCommand:
						varNode.Command = &NodeVarCommand{
							Local:   !Var.Remote,
							Command: strings.Split(Var.Value, " "),
						}
						break
					}
				}

				tokens = append(tokens, Token{
					Key:  keyword,
					Node: varNode,
				})
				break
			}

			continue
		}

		return tokens, errors.New(fmt.Sprintf("Invalid token `%s`", char))
	}

	return tokens, nil
}
