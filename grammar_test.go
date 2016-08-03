package dot

import "testing"
import "github.com/timtadh/data-structures/test"

import (
	"github.com/timtadh/data-structures/errors"
	. "github.com/timtadh/combos"
)

type logCall struct{}

func (l *logCall) Stmt(n *Node) error {
	errors.Logf("DEBUG", "stmt %v", n)
	return nil
}

func (l *logCall) Enter(name string, n *Node) error {
	errors.Logf("DEBUG", "enter %v %v", name, n)
	return nil
}

func (l *logCall) Exit(name string) error {
	errors.Logf("DEBUG", "exit %v", name)
	return nil
}

func TestEmptyGraph(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts")))
	n, err := Parse([]byte(`digraph ast {}`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestEmptyGraphs(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts"))).
			AddKid(NewNode("Graph").
				AddKid(NewNode("GRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts"))).
			AddKid(NewNode("Graph").
				AddKid(NewNode("GRAPH").
					AddKid(NewNode("STRICT"))).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts")))
	n, err := Parse([]byte(`digraph {} graph {} strict graph x {}

	`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphNode(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Node").
						AddKid(NewNode("ID")).
						AddKid(NewNode("Attrs")))))
	n, err := Parse([]byte(`digraph { a }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphNodeWithAttrs(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Node").
						AddKid(NewNode("ID")).
						AddKid(NewNode("Attrs").
							AddKid(NewNode("Attr").
								AddKid(NewNode("ID")).
								AddKid(NewNode("ID")))))))
	n, err := Parse([]byte(`digraph { a [x=y]}`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphNodeWithPortWithAttrs(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Node").
						AddKid(NewNode("ID").
							AddKid(NewNode("Port").
								AddKid(NewNode("ID")).
								AddKid(NewNode("ID")))).
						AddKid(NewNode("Attrs").
							AddKid(NewNode("Attr").
								AddKid(NewNode("ID")).
								AddKid(NewNode("ID")))))),
	)
	n, err := Parse([]byte(`digraph { a:port:e [x=y]}`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphEdge(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Edge").
						AddKid(NewNode("ID")).
						AddKid(NewNode("ID")).
						AddKid(NewNode("Attrs")))))
	n, err := Parse([]byte(`digraph { a-> b }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphEdgeAttr(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Edge").
						AddKid(NewNode("ID")).
						AddKid(NewNode("ID")).
						AddKid(NewNode("Attrs").
							AddKid(NewNode("Attr").
								AddKid(NewNode("ID")).
								AddKid(NewNode("ID"))).
							AddKid(NewNode("Attr").
								AddKid(NewNode("ID")).
								AddKid(NewNode("ID")))))),
	)
	n, err := Parse([]byte(`digraph { 324 -> 234 [calls=1, weight=0.3]}`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphEdges(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Edge").
						AddKid(NewNode("ID")).
						AddKid(NewNode("ID")).
						AddKid(NewNode("Attrs"))).
					AddKid(NewNode("Edge").
						AddKid(NewNode("ID")).
						AddKid(NewNode("ID")).
						AddKid(NewNode("Attrs")))),
	)
	n, err := Parse([]byte(`digraph { a-> b -> c }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphSubGraphEdge(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Edge").
						AddKid(NewNode("SubGraph").
							AddKid(NewNode("ID")).
							AddKid(NewNode("Stmts").
								AddKid(NewNode("Node").
									AddKid(NewNode("ID")).
									AddKid(NewNode("Attrs"))))).
						AddKid(NewNode("SubGraph").
							AddKid(NewNode("ID")).
							AddKid(NewNode("Stmts").
								AddKid(NewNode("Node").
									AddKid(NewNode("ID")).
									AddKid(NewNode("Attrs"))))).
						AddKid(NewNode("Attrs")))),
	)
	n, err := Parse([]byte(`digraph { {a}-> {b} }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphBareAttr(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("Attr").
						AddKid(NewNode("ID")).
						AddKid(NewNode("ID")))),
	)
	n, err := Parse([]byte(`digraph { a=b }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphTypeAttr(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("NodeAttrs").
						AddKid(NewNode("Attr").
							AddKid(NewNode("ID")).
							AddKid(NewNode("ID"))).
						AddKid(NewNode("Attr").
							AddKid(NewNode("ID")).
							AddKid(NewNode("ID"))))),
	)
	n, err := Parse([]byte(`digraph { node [a=b][e=f] }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphTypeAttr2(x *testing.T) {
	t := (*test.T)(x)
	e := NewNode("Graphs").
			AddKid(NewNode("Graph").
				AddKid(NewNode("DIGRAPH")).
				AddKid(NewNode("ID")).
				AddKid(NewNode("Stmts").
					AddKid(NewNode("NodeAttrs").
						AddKid(NewNode("Attr").
							AddKid(NewNode("ID")).
							AddKid(NewNode("ID"))).
						AddKid(NewNode("Attr").
							AddKid(NewNode("ID")).
							AddKid(NewNode("ID"))))),
	)
	n, err := Parse([]byte(`digraph { node [a=b, e=f;] }`))
	t.AssertNil(err)
	t.Assert(n.Equal(e), "expected %v got %v", e, n)
}

func TestGraphNoErr(x *testing.T) {
	t := (*test.T)(x)
	err := StreamParse([]byte(`
		digraph "afp" {
			node [style=filled fillcolor="#f8f8f8"]
			subgraph cluster_L { L [shape=box fontsize=32 label="File: afp\lType: cpu\l27.21ms of 27.21ms total (  100%)\l"] }
			N1 [label="runtime.cgocall\n4.78ms(17.57%)\nof 4.98ms(18.30%)" fontsize=24 shape=box tooltip="runtime.cgocall (4.98ms)"]
			N507 -> N508 [label=" 24.64ms" weight=91 penwidth=5 tooltip="main.main -> main.run (24.64ms)" labeltooltip="main.main -> main.run (24.64ms)"]
		}

	`), &logCall{})
	t.AssertNil(err)
}

