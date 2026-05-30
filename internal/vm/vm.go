package vm

import (
	"fmt"
	"errors"

	// "github.com/avr34/ggvm/internal/vm/core"
	// "github.com/avr34/ggvm/internal/vm/ggwrapper"
	"github.com/avr34/ggvm/internal/vm/mem"
	"github.com/avr34/ggvm/internal/logging"
	"github.com/avr34/ggvm/internal/vm/core"
	"github.com/avr34/ggvm/internal/vm/tokenizer"
)

type VMState struct {
	// This is the main script's filename. Base name without extension.
	MainFilename string

	// List of tokens, number of tokens, and current token being executed.
	Tokens tokenizer.TokenList
	TotalTokens int
	PC     int

	// Stack and memory are both made up of nodes. Stack is a linked list
	// while memory is a string map (for storing variables). Variables are
	// stored as the <token.File>.<varname> in the map to avoid conflicts.
	//
	// Labels are also stored as <token.File>.<labelname>
	Stack  *mem.Stack
	Memory *mem.Memory // Alias for map[string]mem.Node
	LabelMap *mem.LabelMap

	// FirstPass set to true after the first pass function executes. This
	// function will run macro expansion on all import statements, and loads
	// all labels into the LabelMap. Execution will then begin at the label
	// <MainFilename>.main.
	FirstPassComplete bool

	// True when the vm is halted.
	Halted bool

	// List of all imported files, to avoid cyclic imports (running forever, memory leak)
	VisitedFiles []string
}

func errFunc(tag, message string, tok tokenizer.Token) error {
	return fmt.Errorf(logging.ErrLog(tag) + "File %s\tLine %d: " + message, tok.File, tok.Line)
}

func Init(tokens []tokenizer.Token) *VMState {
	memoryMap := make(mem.Memory)
	labelMap := make(mem.LabelMap)
	return &VMState{
		Tokens: tokens,
		TotalTokens: len(tokens),
		PC: 0,
		Stack: &mem.Stack{},
		Memory: &memoryMap,
		LabelMap: &labelMap,
		FirstPassComplete: false,
		Halted: false,
		VisitedFiles: []string{},
	}
}

func (v *VMState) Run() error {
	tag := pTag + "Run]: "

	// Check if there are tokens in token list
	if v.TotalTokens <= 0 {
		return errors.New(logging.ErrLog(tag) + "No tokens in VM context!")
	}

	// Run first pass if not already done
	if !v.FirstPassComplete {
		err := v.FirstPass()
		if err != nil {
			return err
		}
	}

	// Set PC to <MainFilename>.main label
	var err error
	v.PC, err = v.LabelMap.Read(v.MainFilename + ".main")
	if err != nil {
		return err
	}

	// Begin looping.
	for !v.Halted && v.PC <= v.TotalTokens {
		err := v.step()
		if err != nil{
			v.Halted = true
			return err
		}
	}

	return nil
}

// The first pass function will do an initial pass over the token list.
// In doing so, it will run macro expansion on all imports, such that after
// the first pass there are NO import tokens remaining; and it will register
// all labels.
//
// Once the first pass is complete, the Run function will jump to the main
// entrance point and start execution. Running the Run function without
// running this function is fine; Run will execute this function if not
// already done.
//
// Labels and variables are stored in a hashmap, where the keys are strings.
// The keys themselves are <filename>.<var/labelname>, filename doesnt have the
// extension. If you import "drawing/drawcube.ggvm", to jump to the drawing
// function you must run `call "drawcube.<func-name>"`
//
// Since all imports are macro expanded upon discovery, a label's PC value
// should never change, as imports prior to the label would've already been
// recursively expanded already.
func (v *VMState) FirstPass() error {
	// Unrelated, but we need to set MainFilename so gonna set it w/ the first
	// token of the token list.
	if v.MainFilename == "" && v.TotalTokens != 0 {
		v.MainFilename = v.Tokens[0].File
	}

	v.PC = 0

	for v.PC < v.TotalTokens {
		tok := v.Tokens[v.PC]
		v.PC++
		
		switch tok.Type {
		case tokenizer.TokenCommand:
			if tok.Command == core.Import {
				err := v.coreImport(tok)
				if err != nil {
					return err
				}
			}
		case tokenizer.TokenLabel:
			v.LabelMap.Write(tok.File + "." + tok.String, v.PC)
		}
	}

	// Reset VM state for Run() function
	v.PC = 0
	v.FirstPassComplete = true
	v.Halted = false

	return nil
}

func (v *VMState) step() error {
	tag := pTag + "step]: "

	if v.PC >= v.TotalTokens {
		v.Halted = true
		return nil
	}

	tok := v.Tokens[v.PC]
	v.PC++

	switch tok.Type {
	case tokenizer.TokenVar:
		return fmt.Errorf(logging.ErrLog(tag) + "File %s\tLine %d: Should not see variable outside of immediate", tok.File, tok.Line)
	default:
		// If command, int, string, or float
		// Immediate handling is done within execOp
		if err := v.execOp(tok); err != nil {
			return err
		}
	}

	return nil
}

func (v *VMState) execOp(tok tokenizer.Token) error {
	var err error
	if tok.Type != tokenizer.TokenCommand {
		// Push immediate.
		switch tok.Type {
		case tokenizer.TokenString:
			v.Stack.Push(mem.Node{
				Value: tok.String,
				Type: mem.String,
			})
		case tokenizer.TokenFloat:
			v.Stack.Push(mem.Node{
				Value: tok.Float,
				Type: mem.Float,
			})
		case tokenizer.TokenInt:
			v.Stack.Push(mem.Node{
				Value: tok.Int,
				Type: mem.Int,
			})
		}
	} else {
		if tok.Command.OpCode() < 32 {
			err = v.coreExecOp(tok)
			if err != nil {
				return err
			}
		} else {
			err = v.ggExecOp(tok)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
