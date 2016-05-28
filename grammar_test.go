package dot

import "testing"
import "github.com/timtadh/data-structures/test"

func TestEmptyGraph(x *testing.T) {
	t := (*test.T)(x)
	t.Log("Hello")
	n, err := DotParse([]byte(`digraph { 
		/* here be stmt1 */
		/* here be stmt2 */
	}`))
	t.AssertNil(err)
	t.Log(n.Serialize())
}
