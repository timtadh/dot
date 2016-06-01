package dot

import "testing"
import "github.com/timtadh/data-structures/test"

import (
	"fmt"
)

import (
	. "github.com/timtadh/combos"
)

type expecterCallbacks struct{
	t *test.T
	enter, stmt, exit int
	enters []string
	stmts []string
	exits []string
}

func (e *expecterCallbacks) Stmt(n *Node) error {
	defer func() { e.stmt++ }()
	e.t.Log("stmt", n)
	if e.stmt >= len(e.stmts) {
		return fmt.Errorf("Unexpected stmt %v", n)
	}
	if n.Label != e.stmts[e.stmt] {
		return fmt.Errorf("Stmt Expected %v got %v", e.stmts[e.stmt], n)
	}
	return nil
}

func (e *expecterCallbacks) Enter(name string, n *Node) error {
	defer func() { e.enter++ }()
	e.t.Log("enter", n)
	if e.enter >= len(e.enters) {
		return fmt.Errorf("Unexpected enter %v", n)
	}
	if name != e.enters[e.enter] {
		return fmt.Errorf("Enter Expected %v got %v", e.enters[e.enter], n)
	}
	return nil
}

func (e *expecterCallbacks) Exit(name string) error {
	defer func() { e.exit++ }()
	e.t.Log("exit", name)
	if e.exit >= len(e.exits) {
		return fmt.Errorf("Unexpected exit %v", name)
	}
	if name != e.exits[e.exit] {
		return fmt.Errorf("Exit Expected %v got %v", e.exits[e.exit], name)
	}
	return nil
}

func TestStreamEmptyGraph(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{"Graph"},
		exits: []string{"Graph"},
	}
	err := StreamParse([]byte(`digraph ast {}`), e)
	t.AssertNil(err)
}


func TestStreamEdgeNodeSubGraphNoSemi(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{
			"Graph",
			"SubGraph",
		},
		exits: []string{
			"SubGraph",
			"Graph",
		},
		stmts: []string{
			"Node",
			"Edge",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		a -> {b}
	}`), e)
	t.AssertNil(err)
}

func TestStreamEdgeNodeSubGraphSemi(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{
			"Graph",
			"SubGraph",
		},
		exits: []string{
			"SubGraph",
			"Graph",
		},
		stmts: []string{
			"Node",
			"Edge",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		a -> {b};
	}`), e)
	t.AssertNil(err)
}

func TestStreamNodeEdge(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{"Graph"},
		exits: []string{"Graph"},
		stmts: []string{
			"Node",
			"Edge",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		a
		a -> b
	}`), e)
	t.AssertNil(err)
}

func TestStreamNodeEdgeSubgraph(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{"Graph", "SubGraph"},
		exits: []string{"SubGraph", "Graph"},
		stmts: []string{
			"Node",
			"Edge",
			"Node",
			"SubGraph",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		a
		a -> b
		{ a }
	}`), e)
	t.AssertNil(err)
}

func TestStreamSubgraphEdge(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{"Graph", "SubGraph", "SubGraph"},
		exits: []string{"SubGraph", "SubGraph", "Graph"},
		stmts: []string{
			"Node",
			"Node",
			"Edge",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		{ a } -> {b}
	}`), e)
	t.AssertNil(err)
}

func TestStreamNodeEdgeSubgraphSemi(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{"Graph", "SubGraph"},
		exits: []string{"SubGraph", "Graph"},
		stmts: []string{
			"Node",
			"Edge",
			"Node",
			"SubGraph",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		a
		a -> b;
		{ a; }
	}`), e)
	t.AssertNil(err)
}


func TestStreamSubgraphEdgeSemi(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{"Graph", "SubGraph", "SubGraph"},
		exits: []string{"SubGraph", "SubGraph", "Graph"},
		stmts: []string{
			"Node",
			"Node",
			"Edge",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		{ a } -> {b;};
	}`), e)
	t.AssertNil(err)
}

func TestStreamPort(x *testing.T) {
	t := (*test.T)(x)
	e := &expecterCallbacks{
		t: t,
		enters: []string{
			"Graph",
			"SubGraph",
			"SubGraph",
		},
		exits: []string{
			"SubGraph",
			"SubGraph",
			"Graph",
		},
		stmts: []string{
			"Node",
			"Edge",
			"SubGraph",
			"Edge",
		},
	}
	err := StreamParse([]byte(`digraph ast {
		{
			n:a:se [a=x;]
			subgraph foo {
				a:a -> e:a:nw
			}
		} -> y
	}`), e)
	t.AssertNil(err)
}


