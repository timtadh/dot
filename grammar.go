package dot

import (
	"fmt"
	"strings"
)

import (
	"github.com/timtadh/combos"
)

func DotGrammar() *combos.Grammar {
	g := combos.NewGrammar(Tokens, TokenIds)

	g.Start("Graphs")

	g.AddRule("Graphs", 
		g.Alt(
			g.Concat(g.P("Graph"), g.P("Graphs"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					graphs := combos.NewNode("Graphs").AddKid(nodes[0])
					graphs.Children = append(graphs.Children,
						nodes[1].Children...)
					return graphs, nil
				}),
			g.Epsilon(combos.NewNode("Graphs")),
	))

	g.AddRule("Graph",
		g.Alt(
			g.P("GraphStmt"),
			g.P("COMMENT"),
	))

	gEnd := g.Effect()(func(ctx interface{}, nodes ...*combos.Node) error {
		d := ctx.(*DotParser)
		if d.Callbacks != nil {
			return d.Callbacks.Exit("Graph")
		}
		return nil
	})

	g.AddRule("GraphStmt",
		g.Concat(g.P("GraphStart"), g.P("GraphBody"), gEnd)(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
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
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(nodes[2])
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("STRICT"), g.P("GraphType"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("Graph").
						AddKid(nodes[1].AddKid(nodes[0])).
						AddKid(combos.NewNode(d.NextName("graph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("GraphType"), g.P("ID"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(nodes[1])
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("GraphType"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("Graph").
						AddKid(nodes[0]).
						AddKid(combos.NewValueNode("ID", d.NextName("graph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("Graph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
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
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				n := nodes[1]
				n.SetLocation(n.Location().Join(nodes[0].Location(), nodes[2].Location()))
				return n, nil
			})),
	)

	unwrapMultiple := func(n *combos.Node) []*combos.Node {
		if n.Label != "Edges" {
			return []*combos.Node{n}
		}
		nodes := make([]*combos.Node, 0, len(n.Children)-1)
		attrs := n.Get(-1)
		for i := 0; i < len(n.Children)-1; i++ {
			nodes = append(nodes, n.Get(i).AddKid(attrs))
		}
		return nodes
	}

	g.AddRule("Stmts",
		g.Alt(
			g.Concat(g.P("Stmt"), g.P("Stmts"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					if d.Callbacks != nil {
						return combos.NewNode("Stmts"), nil
					} else {
						stmts := nodes[0]
						stmts.Children = append(stmts.Children, nodes[1].Children...)
						return stmts, nil
					}
				}),
			g.Epsilon(combos.NewNode("Stmts")),
	))

	g.AddRule("Stmt",
		g.Concat(g.P("Stmt'"), g.Alt(g.P(";"), g.Epsilon(combos.NewNode("e"))))(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				d := ctx.(*DotParser)
				stmts := combos.NewNode("Stmts")
				for _, stmt := range unwrapMultiple(nodes[0]) {
					stmts.AddKid(stmt)
					if d.Callbacks != nil {
						err := d.Callbacks.Stmt(stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
				}
				return stmts, nil
			}),
	)

	g.AddRule("Stmt'",
		g.Alt(
			// single token choices
			g.P("COMMENT"),
			// rolled choices
			g.P("StmtSubGraphStart"),
			g.P("StmtIDStart"),
			// usual
			g.P("EdgeStmt"), // leading ID, SUBGRAPH, {
			                 // following EdgeReciever --, ->
			g.P("SubGraph"), // leading SUBGRAPH, {
			                 // following SubGraphStart, {
			                 // following SubGraph, Stmt or ;
			g.P("AttrStmt"), // leading ID
			                 // following ID, =
			g.P("NodeStmt"), // leading ID
			                 // following NodeId, ;, [, Stmt
	))

	// EdgeReciever EdgeCont
	//            RHS    AttrList
	//        Edges  RHS
	//    Edge ... Edge
	edgeAction := func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
		edges := nodes[1].Get(0).Get(0)
		rhs := nodes[1].Get(0).Get(1)
		e := combos.NewNode("Edge").AddKid(nodes[0]).AddKid(rhs)
		edges.PrependKid(e)
		edges.AddKid(nodes[1].Get(1))
		return edges, nil
	}

	// NodeId AttrLists
	nodeAction := func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
		n := combos.NewNode("Node").AddKid(nodes[0]).AddKid(nodes[1])
		return n, nil
	}

	g.AddRule("StmtIDStart",
		g.Concat(g.P("ID"), g.Alt(g.P("AttrStmtCont"), g.P("NodeIdCont")))(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				switch nodes[1].Label {
				case "Attrs":
					return nodeAction(ctx, nodes[0], nodes[1])
				case "EdgeCont":
					return edgeAction(ctx, nodes[0], nodes[1])
				case "NodeIdCont":
					port := nodes[1].Get(0)
					id := nodes[0].AddKid(port)
					cont := nodes[1].Get(1)
					switch cont.Label {
					case "Attrs":
						return nodeAction(ctx, id, cont)
					case "EdgeCont":
						return edgeAction(ctx, id, cont)
					default:
						return nil, port.Error("Unexpected node %v", port)
					}
				case "AttrStmtCont":
					stmt := combos.NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[1].Get(1))
					return stmt, nil
				default:
					return nil, nodes[1].Error("Unexpected node %v", nodes[1])
				}
			}),
	)

	g.AddRule("AttrStmtCont",
		g.Concat(g.P("="), g.P("ID"))(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				n := combos.NewNode("AttrStmtCont").AddKid(nodes[0]).AddKid(nodes[1])
				return n, nil
			}),
	)

	g.AddRule("NodeIdCont",
		g.Alt(
			g.Concat(g.P("Port"), g.Alt(g.P("EdgeCont"), g.P("AttrLists")))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					n := combos.NewNode("NodeIdCont").AddKid(nodes[0]).AddKid(nodes[1])
					return n, nil
				}),
			g.P("EdgeCont"),
			g.P("AttrLists"),
	))

	g.AddRule("StmtSubGraphStart",
		g.Concat(g.P("SubGraph"), g.Alt(g.P("EdgeCont"), g.Epsilon(combos.NewNode("e"))))(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				if nodes[1].Label == "e" {
					return nodes[0], nil
				} else {
					return edgeAction(ctx, nodes[0], nodes[1])
				}
			}),
	)


	g.AddRule("AttrStmt",
		g.Alt(
			g.Concat(g.P("ID"), g.P("="), g.P("ID"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					stmt := combos.NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("AttrType"), g.P("AttrLists"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					name := nodes[0].Label[:1] + strings.ToLower(nodes[0].Label[1:])
					stmt := combos.NewNode(name + "Attrs")
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
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					attrs := combos.NewNode("Attrs")
					attrs.Children = append(attrs.Children, nodes[0].Children...)
					attrs.Children = append(attrs.Children, nodes[1].Children...)
					return attrs, nil
				}),
			g.Epsilon(combos.NewNode("Attrs")),
	))

	g.AddRule("AttrList",
		g.Concat(g.P("["), g.P("AttrExprs"), g.P("]"))(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				n := nodes[1]
				n.SetLocation(n.Location().Join(nodes[0].Location(), nodes[2].Location()))
				return n, nil
			}),
	)

	g.AddRule("AttrExprs",
		g.Alt(
			g.Concat(g.P("AttrExpr"), g.P("AttrExprs"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					attrs := combos.NewNode("Attrs").AddKid(nodes[0])
					attrs.Children = append(attrs.Children, nodes[1].Children...)
					return attrs, nil
				}),
			g.Epsilon(combos.NewNode("Attrs")),
	))

	g.AddRule("AttrExpr",
		g.Alt(
			g.Concat(g.P("ID"), g.P("="), g.P("ID"), g.P(";"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					stmt := combos.NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("ID"), g.P("="), g.P("ID"), g.P(","))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					stmt := combos.NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
			g.Concat(g.P("ID"), g.P("="), g.P("ID"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					stmt := combos.NewNode("Attr").
						AddKid(nodes[0]).AddKid(nodes[2])
					return stmt, nil
				}),
	))

	g.AddRule("NodeStmt",
		g.Concat(g.P("NodeId"), g.P("AttrLists"))(nodeAction),
	)

	g.AddRule("NodeId",
		g.Alt(
			g.Concat(g.P("ID"), g.P("Port"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
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
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					port2 := nodes[3].Value.(string)
					switch port2 {
						case "n", "ne", "e", "se", "s", "sw",
						"w", "nw", "c", "_":
							break
						default:
							return nil, nodes[3].Error(fmt.Sprintf("2nd port id must be a dir (n, ne, e, se, s, se, nw, c, _) got : %v", port2))
					}
					n := combos.NewNode("Port").AddKid(nodes[1]).AddKid(nodes[3])
					return n, nil
				}),
			g.Concat(g.P(":"), g.P("ID"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					n := combos.NewNode("Port").AddKid(nodes[1])
					return n, nil
				}),
	))

	g.AddRule("EdgeStmt",
		g.Concat(g.P("EdgeReciever"), g.P("EdgeCont"))(edgeAction),
	)

	g.AddRule("EdgeCont",
		g.Concat(g.P("EdgeRHS"), g.P("AttrLists"))(
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				n := combos.NewNode("EdgeCont").AddKid(nodes[0]).AddKid(nodes[1])
				return n, nil
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
			func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
				if nodes[2] == nil {
					n := combos.NewNode("RHS").
						AddKid(combos.NewNode("Edges")).
						AddKid(nodes[1])
					return n, nil
				} else {
					edges := nodes[2].Get(0)
					rhs := nodes[2].Get(1)
					e := combos.NewNode("Edge").AddKid(nodes[1]).AddKid(rhs)
					edges.PrependKid(e)
					n := combos.NewNode("RHS").
						AddKid(edges).
						AddKid(nodes[1])
					return n, nil
				}
				}),
	)

	g.AddRule("EdgeRHS'",
		g.Alt(
			g.Concat(g.P("EdgeOp"), g.P("EdgeReciever"), g.P("EdgeRHS'"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					if nodes[2] == nil {
						n := combos.NewNode("RHS").
							AddKid(combos.NewNode("Edges")).
							AddKid(nodes[1])
						return n, nil
					} else {
						edges := nodes[2].Get(0)
						rhs := nodes[2].Get(1)
						e := combos.NewNode("Edge").AddKid(nodes[1]).AddKid(rhs)
						edges.PrependKid(e)
						n := combos.NewNode("RHS").
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
		func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
			d := ctx.(*DotParser)
			if d.Callbacks != nil {
				err := d.Callbacks.Exit("SubGraph")
				if err != nil {
					return nil, nodes[0].Error("Stream callback error").Chain(err)
				}
			}
			return nodes[1].AddKid(nodes[2]), nil
		}),
	)

	g.AddRule("SubGraphStart",
		g.Alt(
			g.Concat(g.P("SUBGRAPH"), g.P("ID"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("SubGraph").
						AddKid(nodes[1])
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("SubGraph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat(g.P("SUBGRAPH"))(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("SubGraph").
						AddKid(combos.NewValueNode("ID", d.NextName("subgraph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("SubGraph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
					return stmt, nil
				}),
			g.Concat()(
				func(ctx interface{}, nodes ...*combos.Node) (*combos.Node, *combos.ParseError) {
					d := ctx.(*DotParser)
					stmt := combos.NewNode("SubGraph").
						AddKid(combos.NewValueNode("ID", d.NextName("subgraph")))
					if d.Callbacks != nil {
						err := d.Callbacks.Enter("SubGraph", stmt)
						if err != nil {
							return nil, stmt.Error("Stream callback error").Chain(err)
						}
					}
					return stmt, nil
				}),
	))

	return g
}

