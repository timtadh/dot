package dot

import (
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

var Literals []string
var Keywords []string
var Tokens []string
var TokenIds map[string]int

func init() {
	Literals = []string{
		"[",
		"]",
		"{",
		"}",
		"=",
		",",
		";",
		":",
		"->",
		"--",
	}
	Keywords = []string{
		"NODE",
		"EDGE",
		"GRAPH",
		"DIGRAPH",
		"SUBGRAPH",
		"STRICT",
	}
	Tokens = []string{
		"COMMENT",
		"ID",
	}
	Tokens = append(Tokens, Keywords...)
	Tokens = append(Tokens, Literals...)
	TokenIds = make(map[string]int)
	for i, tok := range Tokens {
		TokenIds[tok] = i
	}
}

func Skip(*lex.Scanner, *machines.Match) (interface{}, error) {
	return nil, nil
}

func AsToken(name string) lex.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(TokenIds[name], string(m.Bytes), m), nil
	}
}

func Literal(s *lex.Scanner, m *machines.Match) (interface{}, error) {
	return s.Token(TokenIds[string(m.Bytes)], string(m.Bytes), m), nil
}

func Lexer() (*lex.Lexer, error) {
	lexer := lex.NewLexer()

	for _, lit := range Literals {
		r := "\\" + strings.Join(strings.Split(lit, ""), "\\")
		lexer.Add([]byte(r), Literal)
	}

	for _, name := range Keywords {
		lexer.Add([]byte(strings.ToLower(name)), AsToken(name))
	}

	lexer.Add([]byte(`//[^\n]*\n?`), AsToken("COMMENT"))
	lexer.Add([]byte(`/\*([^*]|\r|\n|(\*+([^*/]|\r|\n)))*\*+/`), AsToken("COMMENT"))
	lexer.Add([]byte(`([a-z]|[A-Z])([a-z]|[A-Z]|[0-9]|_)*`), AsToken("ID"))
	lexer.Add([]byte(`"([^\\"]|(\\.))*"`), AsToken("ID"))
	lexer.Add([]byte("( |\t|\n|\r)+"), Skip)
	lexer.Add([]byte(`\<`),
		func(scan *lex.Scanner, match *machines.Match) (interface{}, error) {
			str := make([]byte, 0, 10)
			str = append(str, match.Bytes...)
			brackets := 1
			match.EndLine = match.StartLine
			match.EndColumn = match.StartColumn
			for tc := scan.TC; tc < len(scan.Text); tc++ {
				str = append(str, scan.Text[tc])
				match.EndColumn += 1
				if scan.Text[tc] == '\n' {
					match.EndLine += 1
				}
				if scan.Text[tc] == '<' {
					brackets += 1
				} else if scan.Text[tc] == '>' {
					brackets -= 1
				}
				if brackets == 0 {
					match.TC = scan.TC
					scan.TC = tc + 1
					match.Bytes = str
					return AsToken("ID")(scan, match)
				}
			}
			return nil,
				errors.Errorf("unclosed HTML literal starting at %d, (%d, %d)",
					match.TC, match.StartLine, match.StartColumn)
		},
	)
	err := lexer.Compile()
	if err != nil {
		return nil, err
	}
	return lexer, nil
}
