package dot

import (
	"fmt"
)

type DotParser struct {
	nextName int
}

func DotParse(text []byte) (*Node, error) {
	s, err := Lexer.Scanner(text)
	if err != nil {
		return nil, err
	}
	d := &DotParser{nextName: 0}
	n, parseErr := DotGrammar.Parse(s, d)
	if parseErr != nil {
		fmt.Println(parseErr)
		return nil, parseErr
	}
	return n, nil
}

func (d *DotParser) NextName(prefix string) string {
	d.nextName++
	return fmt.Sprintf("%v%d", prefix, d.nextName)
}
