package dot

import "testing"
import "github.com/timtadh/data-structures/test"

func TestEmptyGraph(x *testing.T) {
	t := (*test.T)(x)
	t.Log("Hello")
	n, err := DotParse([]byte(`digraph {
		// stmt comment
		a ["label"=<a node <b>so cool!</b>>]
		a -> b -> c ->d [a = b, ]
		a = s;
		node [a=b e=f
			s=z]
		subgraph { x -> y }
			->
		{ {q -> r} -> z }
	}
	`))
	// n, err := DotParse([]byte(`digraph {
	// 	rankdir=LR;
	// 	wizard=attr
	// 	graph [a=b, c=d; e=f g=h];
	// 	node ["node"=attr]
	// 	a [label="this is a", "wizard"="of the coast"]
	// 	<wacky b>
	// 	c -> d;
	// 	c -> d -> e -> f -> g [asf=sd];
	// 	/* here be stmt1 */
	// 	rankdir=lr;
	// 	a -> b
	// 	a
	// 	a [asdf=sd];
	// 	graph [a=b];
	// 	node [a=b];
	// 	edge [a=b];
	// 	subgraph {
	// 		a -> b;
	// 	} -> subgraph x {
	// 		"whacky"
	// 	}
	// }
	// strict digraph {
	// 	"strict"
	// }
	// `))
	t.AssertNil(err)
	t.Log(n.Serialize())
}
