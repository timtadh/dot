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

func (self *SourceLocation) String() string {
	return fmt.Sprintf("%d:%d-%d:%d",
		self.StartLine, self.StartColumn, self.EndLine, self.EndColumn)
}
func (self *SourceLocation) Join(others ...*SourceLocation) *SourceLocation {
	if self == nil && len(others) > 0 {
		self = others[0]
		others = others[1:]
	} else if self == nil && len(others) == 0 {
		return nil
	}

	min_start_line := self.StartLine
	min_start_col := self.StartColumn
	max_end_line := self.EndLine
	max_end_col := self.EndColumn

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

func (self *Node) Leaf() bool {
	return len(self.Children) == 0
}

func (self *Node) AddKid(kid *Node) *Node {
	if kid != nil {
		self.Children = append(self.Children, kid)
	}
	return self
}

func (self *Node) PrependKid(kid *Node) *Node {
	kids := self.Children
	self.Children = []*Node{kid}
	self.Children = append(self.Children, kids...)
	return self
}

func (self *Node) Kid(label string) *Node {
	for _, kid := range self.Children {
		if kid.Label == label {
			return kid
		}
	}
	return nil
}

func (self *Node) Get(idx int) *Node {
	if idx < 0 {
		idx = len(self.Children) + idx
	}
	return self.Children[idx]
}

func (self *Node) String() string {
	return fmt.Sprintf("(Node %v %d at %v)", self.Label, len(self.Children), self.Location())
}

func (self *Node) Location() *SourceLocation {
	if self == nil {
		return nil
	}
	if self.location != nil {
		return self.location
	}
	locs := make([]*SourceLocation, 0, len(self.Children))
	for _, kid := range self.Children {
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
	self.location = base.Join(others...)
	return self.location
}

func (self *Node) Serialize() string {
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
	nodes := walk(self)
	return strings.Join(nodes, "\n")
}

