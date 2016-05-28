package dot

import "testing"
import "github.com/timtadh/data-structures/test"

func TestEmptyGraph(x *testing.T) {
	t := (*test.T)(x)
	t.Log("Hello")
	n, err := DotParse([]byte("digraph {  }"))
	t.AssertNil(err)
	t.Log(n.Serialize())
}
