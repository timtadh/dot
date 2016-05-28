package dot

import (
	"fmt"
	"strings"
)

func DotGrammar() *Grammar {
	g := NewGrammar(Tokens, TokenIds)

	g.Start("Graphs")

	g.AddRule("Graphs", 
		g.Alt(
			g.Concat(g.P("Graph"), g.P("Graphs"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					graphs := NewNode("Graphs").AddKid(nodes[0])
					graphs.Children = append(graphs.Children,
						nodes[1].Children...)
					return graphs, nil
				}),
			g.Epsilon(NewNode("Graphs")),
	))

	g.AddRule("Graph",
		g.Alt(
			g.P("GraphStmt"),
			g.P("COMMENT"),
	))

	gEnd := g.Effect()(func(ctx interface{}, nodes ...*Node) error {
		d := ctx.(*DotParser)
		if d.Callbacks != nil {
			return d.Callbacks.Exit("Graph")
		}
		return nil
	})

	g.AddRule("GraphStmt",
		g.Concat(g.P("GraphStart"), g.P("GraphBody"), gEnd)(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				stmt := nodes[0].AddKid(nodes[1])
				// force a re-computation of the location of the graph
				// incase the partial action computed it
				stmt.SetLocation(nil)
				return stmt, nil
			}),
	)

	g.AddRule("GraphStart",
		g.Alt(
			g.Concat(g.P("STRICT"), g.P("GraphType"), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(nodes[2])
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("STRICT"), g.P("GraphType"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(NewNode(d.NextName("graph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("GraphType"), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(nodes[1])
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("GraphType"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(NewValueNode("ID", d.NextName("graph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
	))


	g.AddRule("GraphType",
		g.Alt(
			g.P("GRAPH"),
			g.P("DIGRAPH"),
	))

	g.AddRule("GraphBody",
		(g.Concat(g.P("{"), g.P("Stmts"), g.P("}"))(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				n := nodes[1]
				n.SetLocation(n.Location().Join(nodes[0].Location(), nodes[2].Location()))
				return n, nil
			})),
	)

	unwrapMultiple := func(n *Node) []*Node {
		if n.Label != "Edges" {
			return []*Node{n}
		}
		nodes := make([]*Node, 0, len(n.Children)-1)
		attrs := n.Get(-1)
		for i := 0; i < len(n.Children)-1; i++ {
			nodes = append(nodes, n.Get(i).AddKid(attrs))
		}
		return nodes
	}

	g.AddRule("Stmts",
		g.Alt(
			g.Concat(g.P("Stmt"), g.P("Stmts"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					if d.Callbacks != nil {
						return NewNode("Stmts"), nil
					} else {
						stmts := nodes[0]
						stmts.Children = append(stmts.Children, nodes[1].Children...)
						return stmts, nil
					}
				}),
			g.Epsilon(NewNode("Stmts")),
	))

	g.AddRule("Stmt",
		g.Concat(g.Alt(
			g.Concat(g.P("Stmt'"), g.P(";"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					return nodes[0], nil
				}),
			g.P("Stmt'"),
		))(
		func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
			d := ctx.(*DotParser)
			stmts := NewNode("Stmts")
			for _, stmt := range unwrapMultiple(nodes[0]) {
				stmts.AddKid(stmt)
				if d.Callbacks != nil {
					err := d.Callbacks.Stmt(stmt)
					if err != nil {
						return nil, Error("Stream callback error", nil).Chain(err)
					}
				}
			}
			return stmts, nil
		}),
	)

	g.AddRule("Stmt'",
		g.Alt(
			g.P("COMMENT"),
			g.P("EdgeStmt"),
			g.P("SubGraph"),
			g.P("AttrStmt"),
			g.P("NodeStmt"),
	))

	g.AddRule("AttrStmt",
		g.Alt(
			g.Concat(g.P("ID"), g.P("="), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("AttrType"), g.P("AttrLists"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					name := nodes[0].Label[:1] + strings.ToLower(nodes[0].Label[1:])
					stmt := NewNode(name + "Attrs")
					stmt.Children = nodes[1].Children
					return stmt, nil
				}),
	))

	g.AddRule("AttrType",
		g.Alt(
			g.P("NODE"),
			g.P("EDGE"),
			g.P("GRAPH"),
	))

	g.AddRule("AttrLists",
		g.Alt(
			g.Concat(g.P("AttrList"), g.P("AttrLists"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					attrs := NewNode("Attrs")
					attrs.Children = append(attrs.Children, nodes[0].Children...)
					attrs.Children = append(attrs.Children, nodes[1].Children...)
					return attrs, nil
				}),
			g.Epsilon(NewNode("Attrs")),
	))

	g.AddRule("AttrList",
		g.Concat(g.P("["), g.P("AttrExprs"), g.P("]"))(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				n := nodes[1]
				n.SetLocation(n.Location().Join(nodes[0].Location(), nodes[2].Location()))
				return n, nil
			}),
	)

	g.AddRule("AttrExprs",
		g.Alt(
			g.Concat(g.P("AttrExpr"), g.P("AttrExprs"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					attrs := NewNode("Attrs").AddKid(nodes[0])
					attrs.Children = append(attrs.Children, nodes[1].Children...)
					return attrs, nil
				}),
			g.Epsilon(NewNode("Attrs")),
	))

	g.AddRule("AttrExpr",
		g.Alt(
			g.Concat(g.P("ID"), g.P("="), g.P("ID"), g.P(";"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("ID"), g.P("="), g.P("ID"), g.P(","))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("ID"), g.P("="), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					stmt := NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
	))

	g.AddRule("NodeStmt",
		g.Concat(g.P("NodeId"), g.P("AttrLists"))(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				n := NewNode("Node").AddKid(nodes[0]).AddKid(nodes[1])
				return n, nil
			}),
	)

	g.AddRule("NodeId",
		g.Alt(
			g.Concat(g.P("ID"), g.P("Port"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					n := nodes[0].AddKid(nodes[1])
					return n, nil
				}),
			g.P("ID"),
	))

	// TODO: Add Port constratins
	// where second ID in "n", "ne", "e", "se", "s", "sw",
	//                    "w", "nw", "c", "_"
	g.AddRule("Port",
		g.Alt(
			g.Concat(g.P(":"), g.P("ID"), g.P(":"), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					port2 := nodes[3].Value.(string)
					switch port2 {
						case "n", "ne", "e", "se", "s", "sw",
						"w", "nw", "c", "_":
							break
						default:
							return nil, Error(fmt.Sprintf("2nd port id must be a dir got : %v", port2), nodes[3].Token)
					}
					n := NewNode("Port").AddKid(nodes[1]).AddKid(nodes[3])
					return n, nil
				}),
			g.Concat(g.P(":"), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					n := NewNode("Port").AddKid(nodes[1])
					return n, nil
				}),
	))

	g.AddRule("EdgeStmt",
		g.Concat(g.P("EdgeReciever"), g.P("EdgeRHS"), g.P("AttrLists"))(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				// n := NewNode("Edge").AddKid(nodes[1].PrependKid(nodes[0])).AddKid(nodes[2])
				edges := nodes[1].Get(0)
				rhs := nodes[1].Get(1)
				e := NewNode("Edge").AddKid(nodes[0]).AddKid(rhs)
				edges.PrependKid(e)
				edges.AddKid(nodes[2])
				return edges, nil
			}),
	)

	// SubGraph causes extra sg parse
	g.AddRule("EdgeReciever",
		g.Alt(
			g.P("NodeId"),
			g.P("SubGraph"),
	))

	g.AddRule("EdgeRHS",
		g.Concat(g.P("EdgeOp"), g.P("EdgeReciever"), g.P("EdgeRHS'"))(
			func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
				if nodes[2] == nil {
					n := NewNode("RHS").
						AddKid(NewNode("Edges")).
						AddKid(nodes[1])
					return n, nil
				} else {
					edges := nodes[2].Get(0)
					rhs := nodes[2].Get(1)
					e := NewNode("Edge").AddKid(nodes[1]).AddKid(rhs)
					edges.PrependKid(e)
					n := NewNode("RHS").
						AddKid(edges).
						AddKid(nodes[1])
					return n, nil
				}
				}),
	)

	g.AddRule("EdgeRHS'",
		g.Alt(
			g.Concat(g.P("EdgeOp"), g.P("EdgeReciever"), g.P("EdgeRHS'"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					if nodes[2] == nil {
						n := NewNode("RHS").
							AddKid(NewNode("Edges")).
							AddKid(nodes[1])
						return n, nil
					} else {
						edges := nodes[2].Get(0)
						rhs := nodes[2].Get(1)
						e := NewNode("Edge").AddKid(nodes[1]).AddKid(rhs)
						edges.PrependKid(e)
						n := NewNode("RHS").
							AddKid(edges).
							AddKid(nodes[1])
						return n, nil
					}
				}),
			g.Epsilon(nil),
	))

	g.AddRule("EdgeOp",
		g.Alt(
			g.P("->"),
			g.P("--"),
	))

	g.AddRule("SubGraph",
		g.Concat(g.Peek("SUBGRAPH", "{"),
			g.P("SubGraphStart"), g.P("GraphBody"))(
		func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
			d := ctx.(*DotParser)
			if d.Callbacks != nil {
				err := d.Callbacks.Exit("SubGraph")
				if err != nil {
					return nil, Error("Stream callback error", nil).Chain(err)
				}
			}
			return nodes[1].AddKid(nodes[2]), nil
		}),
	)

	g.AddRule("SubGraphStart",
		g.Alt(
			g.Concat(g.P("SUBGRAPH"), g.P("ID"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("SubGraph").
						AddKid(nodes[1])
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("SubGraph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("SUBGRAPH"))(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("SubGraph").
						AddKid(NewValueNode("ID", d.NextName("subgraph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("SubGraph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat()(
				func(ctx interface{}, nodes ...*Node) (*Node, *ParseError) {
					d := ctx.(*DotParser)
					stmt := NewNode("SubGraph").
						AddKid(NewValueNode("ID", d.NextName("subgraph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("SubGraph", stmt)
						if err != nil {
							return nil, Error("Stream callback error", nil).Chain(err)
						}
					}
					return stmt, nil
				}),
	))

	return g
}

