package dot

import (
	"fmt"
)

import (
	lex "github.com/timtadh/lexmachine"
)

type Consumer interface {
	Consume(ctx *Parser) (*Node, *ParseError)
}

type FnConsumer func(ctx *Parser) (*Node, *ParseError)

func (self FnConsumer) Consume(ctx *Parser) (*Node, *ParseError) {
	return self(ctx)
}

type parser struct {
	Tokens []string
	TokenIds map[string]int
	Productions map[string]Consumer
}

type Parser struct {
	p *parser
	s *lex.Scanner
	lastError *ParseError
}

func (p *parser) Epsilon(n *Node) Consumer {
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		return n, nil
	})
}

func (p *parser) Concat(consumers ...Consumer) func(func(...*Node)(*Node, *ParseError)) Consumer {
	return func(action func(...*Node)(*Node, *ParseError)) Consumer {
		// Can't cache the Concat because Indices reuses Index.
		return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
			var nodes []*Node
			var n *Node
			var err *ParseError
			tc := ctx.s.TC
			for _, c := range consumers {
				n, err = c.Consume(ctx)
				if err == nil {
					nodes = append(nodes, n)
				} else {
					ctx.s.TC = tc
					return nil, err
				}
			}
			an, aerr := action(nodes...)
			if aerr != nil {
				ctx.s.TC = tc
				return nil, aerr
			}
			return an, nil
		})
	}
}

func (p *parser) Alt(consumers ...Consumer) Consumer {
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		var err *ParseError = nil
		tc := ctx.s.TC
		for _, c := range consumers {
			ctx.s.TC = tc
			n, e := c.Consume(ctx)
			if e == nil {
				return n, nil
			} else if err == nil || err.Less(e) {
				err = e
			}
		}
		if ctx.lastError == nil || ctx.lastError.Less(err) {
			ctx.lastError = err
		}
		ctx.s.TC = tc
		return nil, err
	})
}

func (p *parser) Consume(token string) Consumer {
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		tc := ctx.s.TC
		t, err, eof := ctx.s.Next()
		if eof {
			ctx.s.TC = tc
			return nil, Error(
				fmt.Sprintf("Ran off the end of the input. expected '%v''", token), nil)
		}
		if err != nil {
			ctx.s.TC = tc
			return nil, Error("Lexer Error", nil).Chain(err)
		}
		tk := t.(*lex.Token)
		if tk.Type == p.TokenIds[token] {
			return NewTokenNode(tk), nil
		}
		ctx.s.TC = tc
		return nil, Error(fmt.Sprintf("Expected %v got %%v", token), tk)
	})
}
