package mem

import (
	"fmt"

	"github.com/fogleman/gg"
	"github.com/avr34/ggvm/internal/logging"
)

var pTag string = "[mem."

type NodeType uint8

type Node struct {
	Value      any
	Type       NodeType
	Next, Prev *Node
}

// For memory use case, ignore Next and Prev pointers.
type Memory map[string]Node

type LabelMap map[string]int

type Stack struct {
	Height uint
	Head   *Node
}

const (
	Int NodeType = iota
	Float
	String
	GGCtxPtr
	ReturnAddress
)

// init
func Init() *Stack {
	var stack = Stack{
		Height: 0,
		Head:   nil,
	}
	return &stack
}

// push
func (a *Stack) Push(b Node) {
	var old *Node
	if a.Height != 0 {
		a.Head.Next = &b
		old = a.Head
	}

	a.Head = &b
	a.Head.Prev = old
	a.Height++
}

// pop
func (a *Stack) Pop() Node {
	if a.Height > 0 {
		ret := *a.Head
		if a.Height > 1 {
			a.Head = a.Head.Prev
			a.Head.Next = nil
		} else {
			a.Head = nil
		}
		a.Height--
		return ret
	}
	return Node{}
}

func (a *Stack) Print() {
	localHeight := a.Height
	head := a.Head

	for localHeight > 0 {
		head.Print()
		head = head.Prev
		localHeight--
	}
}

func (a Node) Print() {
	switch a.Type {
	case Int:
		fmt.Printf("Int:           %d\n", a.Value.(int64))
	case Float:
		fmt.Printf("Float:         %f\n", a.Value.(float64))
	case String:
		fmt.Printf("String:        %s\n", a.Value.(string))
	case GGCtxPtr:
		fmt.Printf("GGCtxPtr:      %p\n", a.Value.(*gg.Context))
	case ReturnAddress:
		fmt.Printf("ReturnAddress: %d\n", a.Value.(int))
	}
}

func (a Node) TypeVal() (NodeType, any) {
	return a.Type, a.Value
}

func (m *Memory) Write(varname string, node Node) {
	(*m)[varname] = node
}

func (m *Memory) Read(varname string) (Node, error) {
	tag := pTag + "Read]: "
	node, exists := (*m)[varname]
	if !exists {
		return Node{}, fmt.Errorf(logging.ErrLog(tag)+"Could not find variable %s", varname)
	}
	return node, nil
}

func (m *LabelMap) Write(label string, index int) {
	(*m)[label] = index
}

func (m *LabelMap) Read(label string) (int, error) {
	tag := pTag + "Read]: "

	index, exists := (*m)[label]
	if !exists {
		return 0, fmt.Errorf(logging.ErrLog(tag)+"Could not find index for label %s", label)
	}

	return index, nil
}

func (a NodeType) String() string {
	switch a {
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case GGCtxPtr:
		return "GGCtxPtr"
	case ReturnAddress:
		return "ReturnAddress"
	}
	return ""
}
