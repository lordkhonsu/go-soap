package dom

// Attribute defines a single attribute for a Node
type Attribute struct {
	Namespace *Namespace
	Name      string
	Value     string
}
