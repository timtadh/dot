package dot

import (
	"fmt"
	"strings"
)

import (
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

var Literals []string // The tokens representing literal strings
var Keywords []string // The keyword tokens
var Tokens []string // All of the tokens (including literals and keywords)
var TokenIds map[string]int // A map from the token names to their int ids
var Lexer *lex.Lexer // The lexer object. Use this to construct a Scanner

// Called a package initialization. Creates the lexer and populates token lists.
func init() {
	initTokens()
	var err error
	Lexer, err = initLexer()
	if err != nil {
		panic(err)
	}
}

func initTokens() {
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

// Creates the lexer object and compiles the NFA.
func initLexer() (*lex.Lexer, error) {
	lexer := lex.NewLexer()

	for _, lit := range Literals {
		r := "\\" + strings.Join(strings.Split(lit, ""), "\\")
		lexer.Add([]byte(r), token(lit))
	}
	for _, name := range Keywords {
		lexer.Add([]byte(strings.ToLower(name)), token(name))
	}

	lexer.Add([]byte(`//[^\n]*\n?`), token("COMMENT"))
	lexer.Add([]byte(`/\*([^*]|\r|\n|(\*+([^*/]|\r|\n)))*\*+/`), token("COMMENT"))
	lexer.Add([]byte(`([a-z]|[A-Z])([a-z]|[A-Z]|[0-9]|_)*`), token("ID"))
	lexer.Add([]byte(`"([^\\"]|(\\.))*"`), token("ID"))
	lexer.Add([]byte("( |\t|\n|\r)+"), skip)
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
					return token("ID")(scan, match)
				}
			}
			return nil,
				fmt.Errorf("unclosed HTML literal starting at %d, (%d, %d)",
					match.TC, match.StartLine, match.StartColumn)
		},
	)

	err := lexer.Compile()
	if err != nil {
		return nil, err
	}
	return lexer, nil
}

func skip(*lex.Scanner, *machines.Match) (interface{}, error) {
	return nil, nil
}

func token(name string) lex.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(TokenIds[name], string(m.Bytes), m), nil
	}
}

