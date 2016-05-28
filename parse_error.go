package dot

import (
	"fmt"
	"strings"
)

import (
	lex "github.com/timtadh/lexmachine"
)

type ParseError struct {
	Reason string
	At *lex.Token
	Chained []error
}

func Error(reason string, at *lex.Token) *ParseError {
	return &ParseError{Reason: reason, At: at}
}

func (p *ParseError) Chain(e error) *ParseError {
	p.Chained = append(p.Chained, e)
	return p
}

func (p *ParseError) Error() string {
	errs := make([]string, 0, len(p.Chained)+1)
	for i := len(p.Chained) - 1; i >= 0; i-- {
		errs = append(errs, p.Chained[i].Error())
	}
	errs = append(errs, p.error())
	return strings.Join(errs, "\n")
}

func (p *ParseError) error() string {
	if p.At == nil && len(p.Chained) == 0 {
		return fmt.Sprintf("Parse Error @ EOS : %v", p.Reason)
	} else {
		return fmt.Sprintf("Parse Error @ %v:%v-%v:%v (%v) : %v",
			p.At.StartLine,
			p.At.StartColumn,
			p.At.EndLine,
			p.At.EndColumn,
			string(p.At.Lexeme),
			p.Reason)
	}
}

func (p *ParseError) Less(o *ParseError) bool {
	if p == nil || o == nil {
		return false
	}
	if p.At == nil || o.At == nil {
		return false
	}
	if p.At.StartLine < o.At.StartLine {
		return true
	} else if p.At.StartLine > o.At.StartLine {
		return false
	}
	if p.At.StartColumn < o.At.StartColumn {
		return true
	} else if p.At.StartColumn > o.At.StartColumn {
		return false
	}
	if p.At.EndLine > o.At.EndLine {
		return true
	} else if p.At.EndLine < o.At.EndLine {
		return false
	}
	if p.At.EndColumn > o.At.EndColumn {
		return true
	} else if p.At.EndColumn < o.At.EndColumn {
		return false
	}
	return false
}
