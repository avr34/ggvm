package core

var pTag string = "[core."

type Inst uint8

const (
	Import Inst = iota
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
	Call
	Ret
	Castint
	Castfloat
	Help
	Print
	Halt
	Unknown
)

var Commands = map[string]Inst{
	"IMPORT":    Import,
	"POP":       Pop,
	"DUP":       Dup,
	"SWAP":      Swap,
	"STORE":     Store,
	"LOAD":      Load,
	"ADD":       Add,
	"SUB":       Sub,
	"MUL":       Mul,
	"DIV":       Div,
	"SQRT":      Sqrt,
	"LT":        Lt,
	"EQ":        Eq,
	"GT":        Gt,
	"JUMP":      Jump,
	"JUMPZ":     Jumpz,
	"CALL":      Call,
	"RET":       Ret,
	"CASTINT":   Castint,
	"CASTFLOAT": Castfloat,
	"HELP":      Help,
	"PRINT":     Print,
	"HALT":      Halt,
	"UNKNOWN":   Unknown,
}

func (op Inst) OpCode() uint8 { return uint8(op) }
func (op Inst) String() string {
	// tag := pTag + "String]: "

	switch op {
	case Import:
		return "IMPORT"
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
	case Call:
		return "CALL"
	case Ret:
		return "RET"
	case Castint:
		return "CASTINT"
	case Castfloat:
		return "CASTFLOAT"
	case Help:
		return "HELP"
	case Print:
		return "PRINT"
	case Halt:
		return "HALT"
	default:
		return "UNKNOWN_CORE"
	}
}

func (a Inst) HasImmediate() (bool, string) {
	// tag := pTag + "HasImmediate]: "

	switch a {
	case Import:
		return true, "str"
	case Store:
		return true, "var"
	case Load:
		return true, "var"
	case Help:
		return true, "str"
	}

	return false, ""
}
