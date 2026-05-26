package tokenizer

import (
	"fmt"
	"errors"
	"strings"
	"unicode"
	"strconv"

	"github.com/avr34/ggvm/internal/logging"
	"github.com/avr34/ggvm/internal/vm/core"
	"github.com/avr34/ggvm/internal/vm/ggwrapper"
)

var pTag string = "[tokenizer."

// The Instruction interface allows us to make both core.Inst and ggwrapper.Inst
// into tokens.
type Instruction interface {
	OpCode() uint8
	String() string
	HasImmediate() (bool, string)
}

// There are 5 types of tokens, and uint8 gives the most minimal way of representing
// that.
type TokenType uint8

// This specifies the token types. There are strings (surrounded by double quotes),
// Floats (contain a decimal point), Ints, and Variable names (begin w/ dollar sign).
// TokenCommand means the token is a command.
const (
	TokenCommand TokenType = iota
	TokenLabel
	TokenString
	TokenFloat
	TokenInt
	TokenVar
)

var AllCommands = make(map[string]Instruction)

// This struct holds a single token. TokenType tells us what type of token it is,
// and the corresponding value and line
type Token struct {
	Type    TokenType
	Command Instruction
	String  string
	Float   float64
	Int     int64
	Varname string
	Line    uint
}

func init() {
	for name, op := range core.Commands {
		AllCommands[name] = op
	}

	for name, op := range ggwrapper.Commands {
		AllCommands[name] = op
	}
}

// Returns the value of a token as empty interface. Must be type asserted when
// used.
func (a Token) Value() any {
	switch a.Type {
	case TokenCommand:
		return a.Command
	case TokenString:
		return a.String
	case TokenFloat:
		return a.Float
	case TokenInt:
		return a.Int
	case TokenVar:
		return a.Varname
	default:
		return nil
	}
}

// State for the tokenizer
type TokenizerState uint8

// Tokenizer states
const (
	StateIdle TokenizerState = iota
	StateComment
	StateCommand
	StateLabel
	StateString
	StateFloat
	StateInt
	StateVar
)

// BNF Grammar:
//
// <idle>      := <label> | <command> | <comment>
// <label>     := <A-Z, a-z>:
// <comment>   := ;<anything>\n
// <command>   := <from list><delimiter> | <from list><delimiter><immediate><delimiter>
// <immediate> := <string> | <int> | <float> | <varname>
// <string>    := "<A-Z, a-z, 0-9, etc>"
// <int>       := <number> | -<number>
// <float>     := <number>.<number> | -<number>.<number>
// <varname>   := $<a-z, 0-9>
func Tokenize(source string) (*[]Token, error) {
	tag := pTag + "Tokenize]: "
	var tokens []Token
	var buffer strings.Builder
	
	var stringStart bool

	var floatStart bool
	var floatDecimal bool
	
	var intStart bool
	
	var stateVarStart bool

	state := StateIdle
	var line uint = 1
	runes := []rune(source)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
	
		// On newline increment line counter.
		if ch == '\n' {
			line++
		}

		switch state {
		case StateIdle:
			if ch == ';' {
				state = StateComment
			} else if unicode.IsLetter(ch) {
				state = StateCommand
				buffer.WriteRune(ch)
			}
		case StateComment:
			// Ignore characters until a newline.
			if ch == '\n' {
				state = StateIdle
			}
		case StateCommand:
			// Read into buffer until delimiter is hit
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' || ch == '_' {
				buffer.WriteRune(ch)
			} else if ch == ':' {
				// Not a command, buffer has label
				state = StateLabel
				i--
			} else if whitespace(ch) {
				// append to tokens and reset buffer. If there's an immediate, check for that.
				command, err := getCommand(buffer.String())
				if err != nil {
					return &tokens, fmt.Errorf(logging.ErrLog(tag) + "Failed to parse command %s", buffer)
				}

				tokens = append(tokens, Token{
					Type: TokenCommand,
					Command: command,
					Line: line,
				})

				buffer.Reset()

				// If it has immediate, go to that. Otherwise go to idle
				if x, immType := hasImmediate(command); x {
					state = immType
				} else {
					state = StateIdle
				}
			}
		case StateLabel:
			if strings.Contains(buffer.String(), ".") {
				return &tokens, errors.New(logging.ErrLog(tag) + "Can't have . in label")
			}
			tokens = append(tokens, Token{
				Type: TokenLabel,
				String: buffer.String(),
				Line: line,
			})
			state = StateIdle
			buffer.Reset()
		case StateString:
			if !stringStart && ch == '"' {
				// Whitespace before the string
				stringStart = true
			} else if stringStart && ch == '"' {
				// String has ended
				tokens = append(tokens, Token{
					Type: TokenString,
					String: buffer.String(),
					Line: line,
				})
				buffer.Reset()
				state = StateIdle
				stringStart = false
			} else if stringStart {
				// Within string
				buffer.WriteRune(ch)
			}
		case StateFloat:
			if !unicode.IsDigit(ch) && ch != '.' && ch != '-' {
				buffer.WriteRune(ch)
				return &tokens, fmt.Errorf(logging.ErrLog(tag) + "Unable to parse %s as float", buffer)
			}

			if !floatStart && (unicode.IsDigit(ch) || ch == '-') {
				// Beginning of float
				buffer.WriteRune(ch)
				floatStart = true
			} else if floatStart && unicode.IsDigit(ch) {
				// Within float
				buffer.WriteRune(ch)
			} else if floatStart && ch == '.' {
				// Decimal point
				buffer.WriteRune(ch)
				floatDecimal = true
			} else if floatStart && floatDecimal && whitespace(ch) {
				// Complete
				val, err := strconv.ParseFloat(buffer.String(), 64)
				if err != nil {
					return &tokens, fmt.Errorf(logging.ErrLog(tag) + "Error parsing float: %w", err)
				}
				tokens = append(tokens, Token{
					Type: TokenFloat,
					Float: val,
					Line: line,
				})
				floatStart, floatDecimal = false, false
				state = StateIdle
				buffer.Reset()
			} else if floatStart && !floatDecimal && whitespace(ch) {
				// Complete, but entered as int. Cast to float64.
				val2, err := strconv.ParseInt(buffer.String(), 10, 64)
				val := float64(val2)
				if err != nil {
					return &tokens, fmt.Errorf(logging.ErrLog(tag) + "Error parsing int: %w", err)
				}
				tokens = append(tokens, Token{
					Type: TokenFloat,
					Float: val,
					Line: line,
				})
				floatStart, floatDecimal = false, false
				state = StateIdle
				buffer.Reset()
			}
		case StateInt:
			if !unicode.IsDigit(ch)  && ch != '-' {
				buffer.WriteRune(ch)
				return &tokens, fmt.Errorf(logging.ErrLog(tag) + "Unable to parse %s as int", buffer)
			}
			
			if !intStart && (unicode.IsDigit(ch) || ch == '-') {
				// Beginning of int
				buffer.WriteRune(ch)
				intStart = true
			} else if intStart && unicode.IsDigit(ch) {
				// Within int
				buffer.WriteRune(ch)
			} else if intStart && whitespace(ch) {
				// Complete
				val, err := strconv.ParseInt(buffer.String(), 10, 64)
				if err != nil {
					return &tokens, fmt.Errorf(logging.ErrLog(tag) + "Error parsing int: %w", err)
				}
				tokens = append(tokens, Token{
					Type: TokenInt,
					Int: val,
					Line: line,
				})
				intStart = false
				state = StateIdle
				buffer.Reset()
			}
		case StateVar:
			if !stateVarStart && ch == '$' {
				// Enter variable. omit the dolla sign
				stateVarStart = true
			} else if stateVarStart && (unicode.IsDigit(ch) || unicode.IsLetter(ch) || ch == '_') {
				// In variable
				buffer.WriteRune(ch)
			} else if stateVarStart && whitespace(ch) {
				// Complete
				tokens = append(tokens, Token{
					Type: TokenVar,
					Varname: buffer.String(),
					Line: line,
				})
				stateVarStart = false
				buffer.Reset()
				state = StateIdle
			}
		}
	}

	return &tokens, nil
}

// Detect whitespace rune, return true or false
func whitespace(a rune) bool {
	if a == ' ' || a == '\t' || a == '\n' || a == '\r' {
		return true
	}
	return false
}

// Gets Instruction from string (case insensitive)
func getCommand(a string) (Instruction, error) {
	a = strings.ToUpper(a)

	// command, err := core.OpCodeStr(a)
	// if err != nil {
	// 	command, err = ggwrapper.OpCodeStr(a)
	// 	if err != nil {
	// 		return command, err
	// 	}
	// }

	command := AllCommands[a]

	return command, nil
}

// Returns whether instruction has immediate, and if so which
// tokenizer state to go into (what immediate to expect).
func hasImmediate(a Instruction) (bool, TokenizerState) {
	b, c := a.HasImmediate()
	var d TokenizerState

	if b {
		switch c {
		case "str":
			d = StateString
		case "int":
			d = StateInt
		case "float":
			d = StateFloat
		case "var":
			d = StateVar
		default:
			// TODO implement better error handling
			return false, 0
		}
	}

	return b, d
}
