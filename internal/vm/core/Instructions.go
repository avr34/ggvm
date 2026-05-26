package core

import (
	"errors"
	"strings"

	"github.com/avr34/ggvm/internal/logging"
)

var pTag string = "[core."

type Inst uint8

const (
	Import Inst = iota
	Pushint
	Pushfloat
	Pushstr
	Pop
	Dup
	Swap
	Store
	Load
	Add
	Sub
	Mul
	Div
	Sqrt
	Lt
	Eq
	Gt
	Jump
	Jumpz
	Jumpi
	Jumpiz
	Castint
	Castfloat
	Help
	Print
	Unknown
)

var Commands = map[string]Inst{
	"IMPORT": Import,
	"PUSHINT": Pushint,
	"PUSHFLOAT": Pushfloat,
	"PUSHSTR": Pushstr,
	"POP": Pop,
	"DUP": Dup,
	"SWAP": Swap,
	"STORE": Store,
	"LOAD": Load,
	"ADD": Add,
	"SUB": Sub,
	"MUL": Mul,
	"DIV": Div,
	"SQRT": Sqrt,
	"LT": Lt,
	"EQ": Eq,
	"GT": Gt,
	"JUMP": Jump,
	"JUMPZ": Jumpz,
	"JUMPI": Jumpi,
	"JUMPIZ": Jumpiz,
	"CASTINT": Castint,
	"CASTFLOAT": Castfloat,
	"HELP": Help,
	"PRINT": Print,
	"UNKNOWN": Unknown,
}

func (op Inst) OpCode() uint8 { return uint8(op) }
func (op Inst) String() string {
	// tag := pTag + "String]: "

	switch op {
	case Import:
		return "IMPORT"
	case Pushint:
		return "PUSHINT"
	case Pushfloat:
		return "PUSHFLOAT"
	case Pushstr:
		return "PUSHSTR"
	case Pop:
		return "POP"
	case Dup:
		return "DUP"
	case Swap:
		return "SWAP"
	case Store:
		return "STORE"
	case Load:
		return "LOAD"
	case Add:
		return "ADD"
	case Sub:
		return "SUB"
	case Mul:
		return "MUL"
	case Div:
		return "DIV"
	case Sqrt:
		return "SQRT"
	case Lt:
		return "LT"
	case Eq:
		return "EQ"
	case Gt:
		return "GT"
	case Jump:
		return "JUMP"
	case Jumpz:
		return "JUMPZ"
	case Jumpi:
		return "JUMPI"
	case Jumpiz:
		return "JUMPIZ"
	case Castint:
		return "CASTINT"
	case Castfloat:
		return "CASTFLOAT"
	case Help:
		return "HELP"
	case Print:
		return "PRINT"
	default:
		return "UNKNOWN_CORE"
	}
}

func (a *Inst) OpCodeStr(b string) error {
	tag := pTag + "OpCodeStr]: "
	b = strings.ToUpper(b)

	switch b {
	case "IMPORT":
		*a = Import
		return nil
	case "PUSHINT":
		*a = Pushint
		return nil
	case "PUSHFLOAT":
		*a = Pushfloat
		return nil
	case "PUSHSTR":
		*a = Pushstr
		return nil
	case "POP":
		*a = Pop
		return nil
	case "DUP":
		*a = Dup
		return nil
	case "SWAP":
		*a = Swap
		return nil
	case "STORE":
		*a = Store
		return nil
	case "LOAD":
		*a = Load
		return nil
	case "ADD":
		*a = Add
		return nil
	case "SUB":
		*a = Sub
		return nil
	case "MUL":
		*a = Mul
		return nil
	case "DIV":
		*a = Div
		return nil
	case "SQRT":
		*a = Sqrt
		return nil
	case "LT":
		*a = Lt
		return nil
	case "EQ":
		*a = Eq
		return nil
	case "GT":
		*a = Gt
		return nil
	case "JUMP":
		*a = Jump
		return nil
	case "JUMPZ":
		*a = Jumpz
		return nil
	case "JUMPI":
		*a = Jumpi
		return nil
	case "JUMPIZ":
		*a = Jumpiz
		return nil
	case "CASTINT":
		*a = Castint
		return nil
	case "CASTFLOAT":
		*a = Castfloat
		return nil
	case "HELP":
		*a = Help
		return nil
	case "PRINT":
		*a = Print
		return nil
	default:
		return errors.New(logging.ErrLog(tag) + "Could not decode " + b)
	}
}

func (a Inst) HasImmediate() (bool, string) {
	// tag := pTag + "HasImmediate]: "

	switch a {
	case Import:
		return true, "str"
	case Pushint:
		return true, "int"
	case Pushfloat:
		return true, "float"
	case Pushstr:
		return true, "str"
	case Store:
		return true, "var"
	case Load:
		return true, "var"
	case Jumpi:
		return true, "str"
	case Jumpiz:
		return true, "str"
	case Help:
		return true, "str"
	}

	return false, ""
}
