package dot

import "testing"
import "github.com/timtadh/data-structures/test"

import (
	"bytes"
	"strings"
)

import (
	lex "github.com/timtadh/lexmachine"
)

func TestHello(x *testing.T) {
	t := (*test.T)(x)
	t.Log("hello")
	l, err := Lexer()
	t.AssertNil(err)
	s, err := l.Scanner([]byte(`
digraph G { a -> b; } <asfd<asdf><a><>asdf>x "asdf\\\\\"" // asdf
strict // asdfa asfwe
/*
	asdf  */ DIGRAPH // asdf`))
	t.AssertNil(err)
	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		t.AssertNil(err)
		token := tok.(*lex.Token)
		t.Log(Tokens[token.Type], token)
	}
}

func match(t *test.T, l *lex.Lexer, text, tokenName string) {
	btext := []byte(text)
	tokenType := TokenIds[tokenName]
	s, err := l.Scanner(btext)
	t.AssertNil(err)
	tok, err, eof := s.Next()
	t.AssertNil(err)
	t.Assert(!eof, "got eof")
	token := tok.(*lex.Token)
	t.Assert(token.Type == tokenType, "wrong type %v != %v", Tokens[tokenType], Tokens[token.Type])
	t.Assert(bytes.Equal(token.Lexeme, btext), "%v != %v : %v", string(token.Lexeme), text, token)
	t.Logf("%v == %v : %v %v", string(token.Lexeme), text, tokenName, token)
	tok, err, eof = s.Next()
	t.Assert(eof, "did not get eof")
	t.AssertNil(err)
	t.Assert(tok == nil, "tok should have been nil %v", tok)
}

func TestLiterals(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	for _, lit := range Literals {
		match(t, l, lit, lit)
	}
}

func TestKeywords(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	for _, keyword := range Keywords {
		match(t, l, strings.ToLower(keyword), keyword)
	}
}

func TestLineComment1(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	match(t, l, "// asdfaefasdf", "COMMENT")
}

func TestLineComment2(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	match(t, l, "// asdfaefasdf\n", "COMMENT")
}

func TestRangeComment(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	match(t, l, "/*// asdfaefasdf\n*/", "COMMENT")
}

func TestID1(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	match(t, l, "asdfa_ASDFwe012", "ID")
	match(t, l, "ASDFasdfa_ASDFwe012", "ID")
}

func TestID2(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	match(t, l, `"asdfaw\wef\"awefwef\\\""`, "ID")
}

func TestID3(x *testing.T) {
	t := (*test.T)(x)
	l, err := Lexer()
	t.AssertNil(err)
	match(t, l, `<asdfa <><awefw><awef><aw>awef>`, "ID")
}

