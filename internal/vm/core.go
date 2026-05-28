package vm

import (
	"os"
	"fmt"
	"math"
	"slices"
	"strings"
	"strconv"
	"path/filepath"

	"github.com/fogleman/gg"

	"github.com/avr34/ggvm/internal/logging"
	"github.com/avr34/ggvm/internal/vm/mem"
	"github.com/avr34/ggvm/internal/vm/tokenizer"
)

var pTag string = "[vm."

func ensureExtension(a string) string {
	ext := filepath.Ext(a)

	if strings.ToLower(ext) != ".ggvm" {
		return a + ".ggvm"
	}

	return a
}

func (v *VMState) getImmediate(tokentype tokenizer.TokenType) (tokenizer.Token, error) {
	tag := pTag + "getImmediate]: "
	var immediate tokenizer.Token
	if v.PC < v.TotalTokens {
		immediate = v.Tokens[v.PC]
		v.PC++
	} else {
		tok := v.Tokens[v.PC - 1]
		return tokenizer.Token{}, errFunc(tag, "PC out of bounds", tok)
	}

	if immediate.Type != tokentype {		
		tok := v.Tokens[v.PC - 1]
		return tokenizer.Token{}, errFunc(tag, "Immediate is not a variable", tok)
	}

	return immediate, nil
}

// This function basically does surgery lol. Runs macro expansion on an import token and its
// immediate, replacing it with tokens of the imported file. In the first pass, all imports
// will be expanded till there are no import statements remaining. Thus, importing is not
// handled at all in the execOp switch statement.
func (v *VMState) coreImport(tok tokenizer.Token) error {
	tag := pTag + "coreImport]: "

	// Get the immediate
	var immediate tokenizer.Token
	if v.PC < v.TotalTokens {
		immediate = v.Tokens[v.PC]
		v.PC++
		if immediate.Type != tokenizer.TokenString {			
			return fmt.Errorf(logging.ErrLog(tag) + "Line %d: Immediate type following import command not a string.\n", tok.Line)
		}
	} else {
		return fmt.Errorf(logging.ErrLog(tag) + "Line %d: End of file reached before import immediate\n", tok.Line)
	}

	// Search for the file first locally, then at ~/ggvm/
	splitpath := strings.Split(immediate.String, "/")
	splitpath[len(splitpath)-1] = ensureExtension(splitpath[len(splitpath)-1])
	joinedpath := filepath.Join(splitpath...)

	// Cannot import main script
	if joinedpath[:len(joinedpath)-5] == v.MainFilename {
		return fmt.Errorf(logging.ErrLog(tag) + "File: %s Line %d: Cannot import yourself\n", tok.File, tok.Line)
	}
	
	// Check for cyclic import
	if slices.Contains(v.VisitedFiles, joinedpath) {
		return fmt.Errorf(logging.ErrLog(tag) + "Line %d: Cyclic import detected. %s imported twice.\n", tok.Line, immediate.String)
	} else {
		v.VisitedFiles = append(v.VisitedFiles, joinedpath)
	}

	// attempt to read locally
	var importstring string
	data, err := os.ReadFile(joinedpath)
	if err != nil {
		// if that failed, attempt to read from ~/ggvm/
		uhd, _ := os.UserHomeDir()
		joinedpath = filepath.Join(uhd, "ggvm", joinedpath)

		data, err = os.ReadFile(joinedpath)
		if err != nil {
			// not found anywhere
			return fmt.Errorf(logging.ErrLog(tag) + "Line %d: File %s not found locally or in ~/ggvm\n", tok.Line, immediate.String)
		}
	}
	importstring = string(data)
	
	// Tokenize
	tokenlistptr, err := tokenizer.Tokenize(importstring, immediate.String)

	// To process the imported file, the IMPORT token and the immediate token are both replaced
	// with the new tokenlist slice, and PC is set to the first token of imported tokens.
	//
	// Current PC:
	//     <old token> <IMPORT> <IMMEDIATE> <next token>
	//                          ^
	//                          v.PC
	// 
	// Decrement PC by 1, copy current v.Tokens into new Tokens up until v.PC.
	// Then append all new tokens. After that, append from old list (v.PC + 2) up until
	// the end of the old token list. Set new token list to v.Tokens. v.PC should now point
	// at the first token of the spliced tokenlist.
	var newtokenlist []tokenizer.Token
	v.PC -= 2

	// copy old list while i < v.PC
	for i, token := range v.Tokens {
		if !(i < v.PC) {
			break
		}

		newtokenlist = append(newtokenlist, token)
	}

	// copy all tokens from imported file
	for _, token := range *tokenlistptr {
		newtokenlist = append(newtokenlist, token)
	}

	// copy remaining tokens from old list
	for i := v.PC + 2; i < v.TotalTokens; i++ {
		newtokenlist = append(newtokenlist, v.Tokens[i])
	}

	// Update total tokens, set v.Tokens to new list and continue
	v.Tokens = newtokenlist
	v.TotalTokens = len(v.Tokens)

	return nil
}

func (v *VMState) corePop() error {
	_ = v.Stack.Pop()
	return nil
}

func (v *VMState) coreDup() error {
	node := v.Stack.Pop()
	v.Stack.Push(node)
	v.Stack.Push(node)
	return nil
}

func (v *VMState) coreSwap() error {
	top := v.Stack.Pop()
	bot := v.Stack.Pop()
	v.Stack.Push(top)
	v.Stack.Push(bot)
	return nil
}

func (v *VMState) coreStore() error {
	// tag := pTag + "coreStore]: "

	immediate, err := v.getImmediate(tokenizer.TokenVar)
	if err != nil {
		return err
	}

	node := v.Stack.Pop()
	v.Memory.Write(immediate.File + "." + immediate.Varname, node) // potential bug here? with next and prev pointers

	return nil
}

func (v *VMState) coreLoad() error {
	// tag := pTag + "coreLoad]: "
	
	immediate, err := v.getImmediate(tokenizer.TokenVar)
	if err != nil {
		return err
	}

	
	node, err := v.Memory.Read(immediate.File + "." + immediate.Varname)
	if err != nil {
		return err
	}
	v.Stack.Push(node)
	return nil
}

func (v *VMState) coreAdd() error {
	tag := pTag + "coreAdd]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot add different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.String:
		val := node1.Value.(string) + node2.Value.(string)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.String,
		})
	case mem.Float:
		val := node1.Value.(float64) + node2.Value.(float64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Float,
		})
	case mem.Int:
		val := node1.Value.(int64) + node2.Value.(int64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Int,
		})
	}
	return nil
}

func (v *VMState) coreSub() error {
	tag := pTag + "coreSub]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot subtract different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	if node1.Type == mem.String {
		return errFunc(tag, "Can't subtract strings", v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.Float:
		val := node1.Value.(float64) - node2.Value.(float64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Float,
		})
	case mem.Int:
		val := node1.Value.(int64) - node2.Value.(int64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Int,
		})
	}
	return nil
}

func (v *VMState) coreMul() error {
	tag := pTag + "coreMul]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot multiply different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	if node1.Type == mem.String {
		return errFunc(tag, "Can't multiply strings", v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.Float:
		val := node1.Value.(float64) * node2.Value.(float64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Float,
		})
	case mem.Int:
		val := node1.Value.(int64) * node2.Value.(int64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Int,
		})
	}
	return nil
}

func (v *VMState) coreDiv() error {
	tag := pTag + "coreDiv]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot divide different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	if node1.Type == mem.String {
		return errFunc(tag, "Can't divide strings", v.Tokens[v.PC - 1])
	}

	if node2.Type == mem.Int && node2.Value.(int64) == 0 {
		return errFunc(tag, "Can't divide by 0", v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.Float:
		val := node1.Value.(float64) / node2.Value.(float64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Float,
		})
	case mem.Int:
		val := node1.Value.(int64) / node2.Value.(int64)
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Int,
		})
	}
	return nil
}

func (v *VMState) coreSqrt() error {
	tag := pTag + "coreSqrt]: "

	node := v.Stack.Pop()

	if node.Type != mem.Float {
		return errFunc(tag, "Can only take sqrt of floats", v.Tokens[v.PC - 1])
	}

	val := math.Sqrt(node.Value.(float64))
	v.Stack.Push(mem.Node{
		Value: val,
		Type: mem.Float,
	})

	return nil
}

func (v *VMState) corePrint() error {
	node := v.Stack.Pop() 

	// restore the node
	v.Stack.Push(node)

	switch node.Type {
	case mem.Int:
		fmt.Printf("%d\n", node.Value.(int64))
	case mem.Float:
		fmt.Printf("%f\n", node.Value.(float64))
	case mem.String:
		fmt.Printf("%s\n", node.Value.(string))
	case mem.GGCtxPtr:
		fmt.Printf("%p\n", node.Value.(*gg.Context))
	case mem.ReturnAddress:
		fmt.Printf("%d\n", node.Value.(int))
	}

	return nil
}

func (v *VMState) coreLt() error {
	tag := pTag + "coreLt]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot compare different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	if node1.Type == mem.String {
		return errFunc(tag, "Cannot compare strings", v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.Float:
		if node1.Value.(float64) < node2.Value.(float64) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	case mem.Int:
		if node1.Value.(int64) < node2.Value.(int64) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	}

	return nil
}

func (v *VMState) coreEq() error {
	tag := pTag + "coreEq]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot compare different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.String:
		if node1.Value.(string) == node2.Value.(string) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	case mem.Float:
		if node1.Value.(float64) == node2.Value.(float64) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	case mem.Int:
		if node1.Value.(int64) == node2.Value.(int64) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	}

	return nil
}

func (v *VMState) coreGt() error {
	tag := pTag + "coreGt]: "

	node2 := v.Stack.Pop()
	node1 := v.Stack.Pop()

	if node1.Type != node2.Type {
		message := fmt.Sprintf("Cannot compare different types %s and %s", node1.Type.String(), node2.Type.String())
		return errFunc(tag, message, v.Tokens[v.PC - 1])
	}

	if node1.Type == mem.String {
		return errFunc(tag, "Cannot compare strings", v.Tokens[v.PC - 1])
	}

	switch node1.Type {
	case mem.Float:
		if node1.Value.(float64) > node2.Value.(float64) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	case mem.Int:
		if node1.Value.(int64) > node2.Value.(int64) {
			v.Stack.Push(mem.Node{
				Value: int64(1),
				Type: mem.Int,
			})
		} else {
			v.Stack.Push(mem.Node{
				Value: int64(0),
				Type: mem.Int,
			})
		}
	}

	return nil
}

// Unconditional jump, no pushing return address to stack.
// Jump label is top of stack.
func (v *VMState) coreJump() error {
	tag := pTag + "coreJump]: "
	
	addrnode := v.Stack.Pop()

	if addrnode.Type != mem.String {
		return errFunc(tag, "Top of stack not a label", v.Tokens[v.PC-1])
	}

	addr, err := v.LabelMap.Read(addrnode.Value.(string))
	if err != nil {
		return err
	}

	v.PC = addr
	return nil
}

// Conditional jump, no pushing return address to stack.
// Jump label is top of stack.
func (v *VMState) coreJumpz() error {
	tag := pTag + "coreJump]: "
	
	addrnode := v.Stack.Pop()
	node := v.Stack.Pop()

	if addrnode.Type != mem.String {
		return errFunc(tag, "Top of stack not a label", v.Tokens[v.PC-1])
	}

	if node.Type == mem.Int && node.Value.(int64) == 0 {
		addr, err := v.LabelMap.Read(addrnode.Value.(string))
		if err != nil {
			return err
		}

		v.PC = addr
	}
	return nil
}

// Jump to label at top of stack, and push return address (PC)
func (v *VMState) coreCall() error {
	tag := pTag + "coreJump]: "

	addrnode := v.Stack.Pop()

	if addrnode.Type != mem.String {
		return errFunc(tag, "Top of stack not a label", v.Tokens[v.PC-1])
	}

	addr, err := v.LabelMap.Read(addrnode.Value.(string))
	if err != nil {
		return err
	}

	// PC is already incremented to next token, so just push.
	v.Stack.Push(mem.Node{
		Value: v.PC,
		Type: mem.ReturnAddress,
	})

	v.PC = addr
	return nil
}

// Pop return address off stack and jump to it unconditionally.
func (v *VMState) coreRet() error {
	tag := pTag + "coreRet]: "

	addrnode := v.Stack.Pop()
	if addrnode.Type != mem.ReturnAddress {
		return errFunc(tag, "Top of stack not a return address", v.Tokens[v.PC-1])
	}

	v.PC = addrnode.Value.(int)
	return nil
}

func (v *VMState) coreCastfloat() error {
	tag := pTag + "coreCastfloat]: "

	node := v.Stack.Pop()
	if node.Type != mem.Int && node.Type != mem.String {
		return errFunc(tag, "Can only cast int or string to float", v.Tokens[v.PC-1])
	}

	switch node.Type {
	case mem.Int:
		v.Stack.Push(mem.Node{
			Value: float64(node.Value.(int64)),
			Type: mem.Float,
		})
	case mem.String:
		val, err := strconv.ParseFloat(node.Value.(string), 64)
		if err != nil {
			message := fmt.Sprintf("Can't cast %s to float", node.Value.(string))
			return errFunc(tag, message, v.Tokens[v.PC-1])
		}
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Float,
		})
	}

	return nil
}

func (v *VMState) coreCastint() error {	
	tag := pTag + "coreCastint]: "

	node := v.Stack.Pop()
	if node.Type != mem.Float && node.Type != mem.String {
		return errFunc(tag, "Can only cast float or string to int", v.Tokens[v.PC-1])
	}

	switch node.Type {
	case mem.Float:
		v.Stack.Push(mem.Node{
			Value: int64(node.Value.(float64)),
			Type: mem.Int,
		})
	case mem.String:
		val, err := strconv.ParseInt(node.Value.(string), 10, 64)
		if err != nil {
			message := fmt.Sprintf("Can't cast %s to int", node.Value.(string))
			return errFunc(tag, message, v.Tokens[v.PC-1])
		}
		v.Stack.Push(mem.Node{
			Value: val,
			Type: mem.Int,
		})
	}

	return nil
}

func (v *VMState) coreCaststring() error {
	tag := pTag + "coreCaststring]: "

	node := v.Stack.Pop()
	if node.Type != mem.Float && node.Type != mem.Int && node.Type != mem.String {
		return errFunc(tag, "Can only cast int or float to string", v.Tokens[v.PC-1])
	}

	switch node.Type {
	case mem.Float:
		v.Stack.Push(mem.Node{
			Value: fmt.Sprintf("%f", node.Value.(float64)),
			Type: mem.String,
		})
	case mem.Int:
		v.Stack.Push(mem.Node{
			Value: fmt.Sprintf("%d", node.Value.(int64)),
			Type: mem.String,
		})
	case mem.String:
		v.Stack.Push(node)
	}

	return nil
}
