package dot

import (
	"fmt"
)

import (
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/data-structures/errors"
)

var TRACE = false

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
	if TRACE {
		errors.Logf("DEBUG", "start lazy %v", l.ProductionName)
	}
	n, e := l.G.Productions[l.ProductionName].Consume(ctx)
	if TRACE {
		name := ""
		if n != nil {
			name = n.Label
		}
		if e == nil {
			errors.Logf("DEBUG", "end lazy %v %v", l.ProductionName, name)
		} else {
			errors.Logf("DEBUG", "fail lazy %v", l.ProductionName)
		}
	}
	return n, e
}

type Parser struct {
	Ctx interface{} // User supplied ctx type passed into Concat Action functions. Optional.
	g *Grammar
	s *lex.Scanner
	lastError *ParseError
	userError *ParseError
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
	n, err := g.Productions[g.StartProduction].Consume(p)
	if err != nil {
		return nil, err
	}

	if p.userError != nil {
		return nil, p.userError
	}
	
	t, serr, eof := s.Next()
	if eof {
		return n, nil
	} else if p.lastError != nil {
		return nil, p.lastError
	} else if serr != nil {
		return nil, Error("Unconsumed Input", nil).Chain(err)
	} else {
		return nil, Error("Unconsumed Input", t.(*lex.Token))
	}
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
					err := Error("Side Effect Error", nil).Chain(doerr)
					ctx.userError = err
					return nil, err
				}
				tok := t.(*lex.Token)
				err := Error("Side Effect Error", tok).Chain(doerr)
				ctx.userError = err
				return nil, err
			}
			n = NewNode("Effect")
			n.Children = nodes
			return n, nil
		})
	}
}

func (g *Grammar) Memoize(c Consumer) Consumer {
	type result struct {
		n *Node
		e *ParseError
		tc int
	}
	var s *lex.Scanner
	var cache map[int]*result
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		if cache == nil || s != ctx.s {
			cache = make(map[int]*result)
			s = ctx.s
		}
		tc := ctx.s.TC
		if res, in := cache[tc]; in {
			// errors.Logf("MEMOIZE", "tc %v, %v, %v", tc, res.n, res.e)
			// node, err := c.Consume(ctx)
			// errors.Logf("MEMOIZE", "check tc %v, %v, %v", tc, node, err)
			ctx.s.TC = res.tc
			return res.n, res.e
		}
		n, e := c.Consume(ctx)
		cache[tc] = &result{n, e, ctx.s.TC}
		return n, e
	})
}

func (g *Grammar) Epsilon(n *Node) Consumer {
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		if TRACE {
			errors.Logf("DEBUG", "epsilon %v", n)
		}
		return n, nil
	})
}

func (g *Grammar) Concat(consumers ...Consumer) func(Action) Consumer {
	return func(action Action) Consumer {
		return (FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
			tc := ctx.s.TC
			nodes, err := g.concat(consumers, ctx)
			if err != nil {
				ctx.s.TC = tc
				return nil, err
			}
			an, aerr := action(ctx.Ctx, nodes...)
			if aerr != nil {
				ctx.s.TC = tc
				ctx.userError = aerr
				return nil, aerr
			}
			return an, nil
		}))
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
	return nodes, nil
}

func (g *Grammar) Alt(consumers ...Consumer) Consumer {
	return (FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		var err *ParseError = nil
		tc := ctx.s.TC
		always := false
		for _, c := range consumers {
			ctx.s.TC = tc
			n, e := c.Consume(ctx)
			if e == nil {
				return n, nil
			} else if err == nil {
				err = e
			} else if e.Less(err) {
				// err = err.Chain(e)
				err = err
			} else {
				// err = e.Chain(err)
				err = e
			}
			if ctx.lastError == nil || always {
				ctx.lastError = err
			} else if ctx.lastError.Less(err) {
				always = true
				ctx.lastError = err
			}
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
		return nil, Error(fmt.Sprintf("Expected %v", token), tk)
	})
}

func (g *Grammar) Peek(tokens ...string) Consumer {
	return FnConsumer(func(ctx *Parser) (*Node, *ParseError) {
		tc := ctx.s.TC
		t, err, eof := ctx.s.Next()
		ctx.s.TC = tc
		if eof {
			return nil, Error(
				fmt.Sprintf("Ran off the end of the input. expected '%v''", token), nil)
		}
		if err != nil {
			return nil, Error("Lexer Error", nil).Chain(err)
		}
		tk := t.(*lex.Token)
		for _, token := range tokens {
			if tk.Type == g.TokenIds[token] {
				return nil, nil
			}
		}
		return nil, Error(fmt.Sprintf("Expected one of %v", tokens), tk)
	})
}
