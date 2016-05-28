package dot


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

	g.AddRule("GraphStmt",
		g.Alt(
			g.Concat(g.P("GraphType"), g.P("GraphBody"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(NewNode(d.NextName("graph"))).
						AddKid(nodes[1])
					return stmt, nil
				}),
			g.Concat(g.P("GraphType"), g.P("ID"), g.P("GraphBody"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(nodes[1]).
						AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("STRICT"), g.P("GraphType"), g.P("GraphBody"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(NewNode(d.NextName("graph"))).
						AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("STRICT"), g.P("GraphType"), g.P("ID"), g.P("GraphBody"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(nodes[2]).
						AddKid(nodes[3])
					return stmt, nil
				}),
	))


	g.AddRule("GraphType",
		g.Alt(
			g.P("GRAPH"),
			g.P("DIGRAPH"),
	))

	g.AddRule("GraphBody",
		g.Alt(
			g.Concat(g.P("{"), g.P("}"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					n := NewNode("Graph")
					n.SetLocation(nodes[0].Location().Join(nodes[1].Location()))
					return n, nil
				}),
	))

	DotGrammar = g
}
