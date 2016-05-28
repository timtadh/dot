package dot

import (
	"github.com/timtadh/data-structures/errors"
)

var DotGrammar *Grammar

func initGrammar() {
	g := NewGrammar(Tokens, TokenIds)
	g.Start("Graphs")

	g.AddRule("Graphs", 
		g.Alt(
			g.Concat(g.P("Graph"), g.P("Graphs"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					graphs := NewNode("Graphs").AddKid(nodes[0])
					graphs.Children = append(graphs.Children, nodes[1].Children...)
					return graphs, nil
				}),
			g.Concat(g.P("Graph"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					return NewNode("Graphs").AddKid(nodes[0]), nil
				}),
	))

	g.AddRule("Graph",
		g.Alt(
			g.P("GraphStmt"),
			g.P("COMMENT"),
	))

	// TODO: The effect needs to capture the tokens before it. However,
	// this is an effective demonstration of where we could insert a callback
	// informing user code of the start of a new graph statment.
	gStart := g.Effect()(func(ctx interface{}, nodes ...*Node) error {
		errors.Logf("DEBUG", "graph start")
		return nil
	})
	
	// TODO: This effect needs to call back to indicate the end of the graph.
	gEnd := g.Effect()(func(ctx interface{}, nodes ...*Node) error {
		errors.Logf("DEBUG", "graph end")
		return nil
	})

	g.AddRule("GraphStmt",
		g.Alt(
			g.Concat(g.P("GraphType"), gStart, g.P("GraphBody"), gEnd)(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(NewNode(d.NextName("graph"))).
						AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("GraphType"), g.P("ID"), gStart, g.P("GraphBody"), gEnd)(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(nodes[1]).
						AddKid(nodes[3])
					return stmt, nil
				}),
			g.Concat(g.P("STRICT"), g.P("GraphType"), gStart, g.P("GraphBody"), gEnd)(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(NewNode(d.NextName("graph"))).
						AddKid(nodes[3])
					return stmt, nil
				}),
			g.Concat(g.P("STRICT"), g.P("GraphType"), g.P("ID"), gStart, g.P("GraphBody"), gEnd)(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(nodes[2]).
						AddKid(nodes[4])
					return stmt, nil
				}),
	))


	g.AddRule("GraphType",
		g.Alt(
			g.P("GRAPH"),
			g.P("DIGRAPH"),
	))

	g.AddRule("GraphBody",
		g.Concat(g.P("{"), g.P("Stmts"), g.P("}"))(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				n := nodes[1]
				n.SetLocation(n.Location().Join(nodes[0].Location(), nodes[2].Location()))
				return n, nil
			}),
	)

	g.AddRule("Stmts",
		g.Alt(
			g.Concat(g.P("Stmt"), g.P("Stmts"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmts := NewNode("Stmts").AddKid(nodes[0])
					stmts.Children = append(stmts.Children, nodes[1].Children...)
					return stmts, nil
				}),
			g.Concat(g.P("Stmt"), g.P(";"), g.P("Stmts"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmts := NewNode("Stmts").AddKid(nodes[0])
					stmts.Children = append(stmts.Children, nodes[2].Children...)
					return stmts, nil
				}),
			g.Epsilon(NewNode("Stmts")),
	))

	g.AddRule("Stmt", g.P("COMMENT"))

	DotGrammar = g
}
