package dot

import (
	"fmt"
)

type Callbacks interface {
	Stmt(*Node) error
	Enter(name string, n *Node) error
	Exit(name string) error
}

type DotParser struct {
	nextName int
	Callbacks Callbacks
}

func NewDotParser(c Callbacks) *DotParser {
	return &DotParser{Callbacks: c}
}

func StreamParse(text []byte, call Callbacks) error {
	_, err := dotParse(text, call)
	return err
}

func Parse(text []byte) (*Node, error) {
	return dotParse(text, nil)
}

func dotParse(text []byte, call Callbacks) (*Node, error) {
	s, err := Lexer.Scanner(text)
	if err != nil {
		return nil, err
	}
	n, parseErr := DotGrammar().Parse(s, NewDotParser(call))
	if parseErr != nil {
		return nil, parseErr
	}
	return n, nil
}

func (d *DotParser) NextName(prefix string) string {
	d.nextName++
	return fmt.Sprintf("%v%d", prefix, d.nextName)
}
