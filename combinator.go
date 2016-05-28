package dot

import (
	"fmt"
)

import (
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/data-structures/errors"
)

type Action func(ctx interface{}, nodes ...*Node) (*Node, *ParseError)

type Consumer interface {
	Consume(ctx *Parser) (*Node, *ParseError)
}

type FnConsumer func(ctx *Parser) (*Node, *ParseError)

func (self FnConsumer) Consume(ctx *Parser) (*Node, *ParseError) {
	return self(ctx)
}

type LazyConsumer struct {
	G *Grammar
	ProductionName string
}

func (l *LazyConsumer) Consume(ctx *Parser) (*Node, *ParseError) {
	errors.Logf("DEBUG", "lazy %v", l.ProductionName)
	return l.G.Productions[l.ProductionName].Consume(ctx)
}

type Parser struct {
	Ctx interface{} // User supplied ctx type passed into Concat Action functions. Optional.
	g *Grammar
	s *lex.Scanner
	lastError *ParseError
}


type Grammar struct {
	Tokens []string
	TokenIds map[string]int
	Productions map[string]Consumer
	StartProduction string
}

func NewGrammar(tokens []string, tokenIds map[string]int) *Grammar {
	g := &Grammar{
		Tokens: tokens,
		TokenIds: tokenIds,
		Productions: make(map[string]Consumer),
	}
	for _, token := range Tokens {
		g.AddRule(token, g.Consume(token))
	}
	return g
}

func (g *Grammar) Parse(s *lex.Scanner, parserCtx interface{}) (*Node, *ParseError) {
	p := &Parser{
		Ctx: parserCtx,
		g: g,
		s: s,
	}
	return g.Productions[g.StartProduction].Consume(p)
}

func (g *Grammar) Start(name string) *Grammar {
	g.StartProduction = name
	return g
}

func (g *Grammar) AddRule(name string, c Consumer) *Grammar {
	g.Productions[name] = c
	return g
}

func (g *Grammar) P(productionName string) Consumer {
	return &LazyConsumer{G: g, ProductionName: productionName}
}

func (g *Grammar) Memoize(c Consumer) Consumer {
	type result struct {
		n *Node
		e *ParseError
	}
	cache := make(map[int]*result)
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		tc := ctx.s.TC
		if res, in := cache[tc]; in {
			return res.n, res.e
		}
		n, e := c.Consume(ctx)
		cache[tc] = &result{n, e}
		return n, e
	})
}

func (g *Grammar) Effect(consumers ...Consumer) func(do func(interface{}, ...*Node) error) Consumer {
	return func(do func(interface{}, ...*Node) error) Consumer {
		return FnConsumer(func(ctx *Parser) (n *Node, err *ParseError) {
			tc := ctx.s.TC
			nodes, err := g.concat(consumers, ctx)
			if err != nil {
				ctx.s.TC = tc
				return nil, err
			}
			doerr := do(ctx.Ctx, nodes...)
			if doerr != nil {
				ctx.s.TC = tc
				t, _, _ := ctx.s.Next()
				if t == nil {
					return nil, Error("Side Effect Error", nil).Chain(doerr)
				}
				tok := t.(*lex.Token)
				return nil, Error("Side Effect Error", tok).Chain(doerr)
			}
			n = NewNode("Effect")
			n.Children = nodes
			return n, nil
		})
	}
}

func (g *Grammar) Epsilon(n *Node) Consumer {
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		return n, nil
	})
}

func (g *Grammar) Concat(consumers ...Consumer) func(Action) Consumer {
	return func(action Action) Consumer {
		// Can't cache the Concat because Indices reuses Index.
		return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
			tc := ctx.s.TC
			nodes, err := g.concat(consumers, ctx)
			if err != nil {
				ctx.s.TC = tc
				return nil, err
			}
			an, aerr := action(ctx.Ctx, nodes...)
			if aerr != nil {
				ctx.s.TC = tc
				return nil, aerr
			}
			return an, nil
		})
	}
}

func (g *Grammar) concat(consumers []Consumer, ctx *Parser) ([]*Node, *ParseError) {
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
	return nodes, err
}

func (g *Grammar) Alt(consumers ...Consumer) Consumer {
	return g.Memoize(FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
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
	}))
}

func (g *Grammar) Consume(token string) Consumer {
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
		if tk.Type == g.TokenIds[token] {
			return NewTokenNode(tk), nil
		}
		ctx.s.TC = tc
		return nil, Error(fmt.Sprintf("Expected %v got %%v", token), tk)
	})
}
