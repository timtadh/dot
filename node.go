package dot

import (
	"fmt"
	"strings"
)

import (
	lex "github.com/timtadh/lexmachine"
)


type SourceLocation struct {
	StartLine, StartColumn, EndLine, EndColumn int
}

func (n *SourceLocation) String() string {
	return fmt.Sprintf("%d:%d-%d:%d",
		n.StartLine, n.StartColumn, n.EndLine, n.EndColumn)
}
func (n *SourceLocation) Join(others ...*SourceLocation) *SourceLocation {
	if n == nil && len(others) > 0 {
		n = others[0]
		others = others[1:]
	} else if n == nil && len(others) == 0 {
		return nil
	}

	min_start_line := n.StartLine
	min_start_col := n.StartColumn
	max_end_line := n.EndLine
	max_end_col := n.EndColumn

	for _, o := range others {
		if o.StartLine < min_start_line {
			min_start_line = o.StartLine
			min_start_col = o.StartColumn
		} else if o.StartLine == min_start_line && o.StartColumn < min_start_col {
			min_start_col = o.StartColumn
		}
		if o.EndLine > max_end_line {
			max_end_line = o.EndLine
			max_end_col = o.EndColumn
		} else if o.EndLine == max_end_line && o.EndColumn > max_end_col {
			max_end_col = o.EndColumn
		}
	}

	return &SourceLocation{
		StartLine:min_start_line, StartColumn:min_start_col,
		EndLine:max_end_line, EndColumn:max_end_col,
	}
}

type Node struct {
	Label    string
	Value    interface{}
	Children []*Node
	location *SourceLocation
}

func NewNode(label string) *Node {
	return &Node{
		Label:    label,
		Value:    nil,
		Children: make([]*Node, 0, 5),
	}
}

func NewValueNode(label string, value interface{}) *Node {
	return &Node{
		Label:    label,
		Value:    value,
		Children: make([]*Node, 0, 5),
	}
}

func NewTokenNode(tok *lex.Token) *Node {
	return &Node{
		Label: Tokens[tok.Type],
		Value: tok.Value,
		location: &SourceLocation{
			StartLine: tok.StartLine,
			StartColumn: tok.StartColumn,
			EndLine: tok.EndLine,
			EndColumn: tok.EndColumn,
		},
	}
}

func (n *Node) Leaf() bool {
	return len(n.Children) == 0
}

func (n *Node) AddKid(kid *Node) *Node {
	if kid != nil {
		n.Children = append(n.Children, kid)
	}
	return n
}

func (n *Node) PrependKid(kid *Node) *Node {
	kids := n.Children
	n.Children = []*Node{kid}
	n.Children = append(n.Children, kids...)
	return n
}

func (n *Node) Kid(label string) *Node {
	for _, kid := range n.Children {
		if kid.Label == label {
			return kid
		}
	}
	return nil
}

func (n *Node) Get(idx int) *Node {
	if idx < 0 {
		idx = len(n.Children) + idx
	}
	return n.Children[idx]
}

func (n *Node) String() string {
	return fmt.Sprintf("(Node %v %d at %v)", n.Label, len(n.Children), n.Location())
}

func (n *Node) SetLocation(sl *SourceLocation) {
	n.location = sl
}

func (n *Node) Location() *SourceLocation {
	if n == nil {
		return nil
	}
	if n.location != nil {
		return n.location
	}
	locs := make([]*SourceLocation, 0, len(n.Children))
	for _, kid := range n.Children {
		kl := kid.Location()
		if kl != nil {
			locs = append(locs, kl)
		}
	}
	if len(locs) == 0 {
		return nil
	} else if len(locs) == 1 {
		return locs[0]
	}
	base := locs[0]
	others := locs[1:]
	n.location = base.Join(others...)
	return n.location
}

func (n *Node) Serialize() string {
	fmt_node := func(n *Node) string {
		s := ""
		loc := n.Location()
		if n.Value != nil && loc != nil {
			s = fmt.Sprintf(
				"%d:%s (%v) @ %v",
				len(n.Children),
				n.Label,
				n.Value,
				loc,
			)
		} else if n.Value != nil {
			s = fmt.Sprintf(
				"%d:%s (%v)",
				len(n.Children),
				n.Label,
				n.Value,
			)
		} else if loc != nil {
			s = fmt.Sprintf(
				"%d:%s @ %v",
				len(n.Children),
				n.Label,
				loc,
			)
		} else {
			s = fmt.Sprintf(
				"%d:%s",
				len(n.Children),
				n.Label,
			)
		}
		return s
	}
	walk := func(node *Node) (nodes []string) {
		type entry struct {
			n *Node
			i int
		}
		type node_stack []*entry
		pop := func(stack node_stack) (node_stack, *entry) {
			if len(stack) <= 0 {
				return stack, nil
			} else {
				return stack[0 : len(stack)-1], stack[len(stack)-1]
			}
		}

		stack := make(node_stack, 0, 10)
		stack = append(stack, &entry{node, 0})

		for len(stack) > 0 {
			var c *entry
			stack, c = pop(stack)
			if c.i == 0 {
				nodes = append(nodes, fmt_node(c.n))
			}
			if c.i < len(c.n.Children) {
				kid := c.n.Children[c.i]
				stack = append(stack, &entry{c.n, c.i + 1})
				stack = append(stack, &entry{kid, 0})
			}
		}
		return nodes
	}
	nodes := walk(n)
	return strings.Join(nodes, "\n")
}

