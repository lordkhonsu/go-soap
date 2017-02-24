package dom

// NodeList defines a list of nodes
type NodeList struct {
	nodes []*Node
}

// First returns the first node in the list
func (n *NodeList) First() *Node {
	if n.nodes == nil || len(n.nodes) == 0 {
		return &Node{}
	}
	return n.nodes[0]
}

// All returns all Nodes from this list
func (n *NodeList) All() []*Node {
	if n.nodes == nil {
		return []*Node{}
	}
	return n.nodes
}

// Len returns the length of this list
func (n *NodeList) Len() int {
	if n.nodes == nil {
		return 0
	}
	return len(n.nodes)
}

// Append adds a new Node to the list
func (n *NodeList) Append(node *Node) {
	n.nodes = append(n.nodes, node)
}

// AppendList appends all Nodes from another list to this list
func (n *NodeList) AppendList(list *NodeList) {
	for _, node := range list.All() {
		n.Append(node)
	}
}

// ClearAll removes all Nodes from this list
func (n *NodeList) ClearAll() {
	n.nodes = []*Node{}
}
