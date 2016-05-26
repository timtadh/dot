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

func TestIntegration(x *testing.T) {
	t := (*test.T)(x)
	t.Log("hello")
	s, err := Lexer.Scanner([]byte(`
digraph G { a -> b; } <asfd<asdf><a><>asdf>x "asdf\\\\\"" // asdf
strict // asdfa asfwe
/*
	asdf  */ DIGRAPH // asdf`))
	t.AssertNil(err)
	expected := []int{
		TokenIds["DIGRAPH"],
		TokenIds["ID"],
		TokenIds["{"],
		TokenIds["ID"],
		TokenIds["->"],
		TokenIds["ID"],
		TokenIds[";"],
		TokenIds["}"],
		TokenIds["ID"],
		TokenIds["ID"],
		TokenIds["ID"],
		TokenIds["COMMENT"],
		TokenIds["STRICT"],
		TokenIds["COMMENT"],
		TokenIds["COMMENT"],
		TokenIds["ID"],
		TokenIds["COMMENT"],
	}
	i := 0
	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		t.AssertNil(err)
		token := tok.(*lex.Token)
		t.Log(i, Tokens[expected[i]], Tokens[token.Type], token)
		t.Assert(token.Type == expected[i], "got %v expected %v", Tokens[token.Type], Tokens[expected[i]])
		i++
	}
}

func match(t *test.T, text, tokenName string) {
	btext := []byte(text)
	tokenType := TokenIds[tokenName]
	s, err := Lexer.Scanner(btext)
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

func not_match(t *test.T, matches int, text, tokenName string) {
	btext := []byte(text)
	s, err := Lexer.Scanner(btext)
	t.AssertNil(err)
	for i := 0; i < matches; i++ {
		tok, err, eof := s.Next()
		t.Assert(!eof, "got eof")
		t.AssertNil(err)
		t.Log(tok)
	}
	tok, err, eof := s.Next()
	t.Assert(!eof, "got eof")
	t.Assert(tok == nil, "tok should have been nil %v", tok)
	t.Assert(err != nil, "got nil err when there should have been an error")
	t.Logf("As expected for '%v' which looks like a %v\n%v", text, tokenName, err)
}

func TestLiterals(x *testing.T) {
	t := (*test.T)(x)
	for _, lit := range Literals {
		match(t, lit, lit)
	}
}

func TestKeywords(x *testing.T) {
	t := (*test.T)(x)
	for _, keyword := range Keywords {
		match(t, strings.ToLower(keyword), keyword)
	}
}

func TestLineComment1(x *testing.T) {
	t := (*test.T)(x)
	match(t, "// asdfaefasdf", "COMMENT")
}

func TestLineComment2(x *testing.T) {
	t := (*test.T)(x)
	match(t, "// asdfaefasdf\n", "COMMENT")
}

func TestRangeComment(x *testing.T) {
	t := (*test.T)(x)
	match(t, "/*// asdfaefasdf\n*/", "COMMENT")
}

func TestID1(x *testing.T) {
	t := (*test.T)(x)
	match(t, "asdfa_ASDFwe012", "ID")
	match(t, "ASDFasdfa_ASDFwe012", "ID")
}

func TestID2(x *testing.T) {
	t := (*test.T)(x)
	match(t, `"asdfaw\wef\"awefwef\\\""`, "ID")
}

func TestID3(x *testing.T) {
	t := (*test.T)(x)
	match(t, `<asdfa <><awefw><awef><aw>awef>`, "ID")
}

func TestNotID1(x *testing.T) {
	t := (*test.T)(x)
	not_match(t, 0, "_asdf", "ID")
}

func TestNotID2(x *testing.T) {
	t := (*test.T)(x)
	not_match(t, 1, "as342*df", "ID")
}

func TestNotID3(x *testing.T) {
	t := (*test.T)(x)
	not_match(t, 0, `"asdf\"`, "ID")
}

func TestNotID4(x *testing.T) {
	t := (*test.T)(x)
	not_match(t, 0, `<"asdf<\>"`, "ID")
}
