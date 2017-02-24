package dom

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Node defines a single element in our DOM
type Node struct {
	Namespace        *Namespace
	Document         *Document
	Parent           *Node
	Name             string
	Attributes       []*Attribute
	Children         NodeList
	Value            Value
	Exists           bool
	NamespaceMapping map[string]*Namespace
}

// XML converts this Node into its XML expression
func (n *Node) XML() string {
	return n.xml(0)
}

// @todo lots of escaping
// @todo CDATA
func (n *Node) xml(indent int) string {
	result := ""
	indentation := strings.Repeat(" ", indent)

	result += fmt.Sprintf("%s", indentation)

	// opening tag
	result += "<"
	if n.Namespace != nil && n.Namespace.Abbreviation != "" {
		result += n.Namespace.Abbreviation + ":"
	}
	result += n.Name

	// attributes
	for _, attr := range n.Attributes {
		result += " "
		if attr.Namespace != nil && attr.Namespace.Abbreviation != "" {
			result += attr.Namespace.Abbreviation + ":"
		}
		result += attr.Name + "=\"" + attr.Value + "\""
	}

	// namespaces
	for _, ns := range n.NamespaceMapping {
		result += " xmlns"
		if ns.Abbreviation != "" {
			result += ":" + ns.Abbreviation
		}
		result += "=\"" + ns.Name + "\""
	}

	// short closing tag
	if n.Value.value == nil && n.Children.Len() == 0 {
		result += "/>\n"
		return result
	}

	result += ">" + n.String()

	// children
	if n.Children.Len() > 0 {
		result += "\n"
		for _, child := range n.Children.All() {
			result += child.xml(indent + 2)
		}
		result += indentation
	}

	// closing tag
	result += "</"
	if n.Namespace != nil && n.Namespace.Abbreviation != "" {
		result += n.Namespace.Abbreviation + ":"
	}
	result += n.Name + ">\n"

	return result
}

// UnmarshalXML implements the xml.Unmarshaler interface for decoding
func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// init Node
	n.Exists = true
	n.Name = start.Name.Local
	n.Attributes = []*Attribute{}

	lateBinding := make([]xml.Attr, 0, len(start.Attr))
	for _, attr := range start.Attr {
		if attr.Name.Space == "xmlns" {
			n.RegisterNS(attr.Value, attr.Name.Local)
		} else if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
			n.RegisterNS(attr.Value, "")
		} else {
			lateBinding = append(lateBinding, attr)
		}
	}

	for _, attr := range lateBinding {
		n.SetAttribute(attr.Name.Space, attr.Name.Local, attr.Value)
	}

	// save namespace
	n.Namespace = n.ResolveNS(start.Name.Space)

	// traverse
	end := start.End()
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		if token == end {
			return nil
		}

		switch t := token.(type) {
		case xml.StartElement:
			child := &Node{Document: n.Document, Parent: n}
			d.DecodeElement(child, &t)
			n.Children.Append(child)

		case xml.CharData:
			n.Value.value = string(t)

		case xml.EndElement:
			return fmt.Errorf("Unexpected xml.EndElement")
		}
	}
}

// SetAttribute sets a given attribute to the current Node
func (n *Node) SetAttribute(namespace string, name string, value string) {
	// replace
	for _, attr := range n.Attributes {
		if attr.Name == name {
			attr.Value = value
			return
		}
	}

	// append (default ns)
	if namespace == "" {
		n.Attributes = append(n.Attributes, &Attribute{
			Namespace: nil,
			Name:      name,
			Value:     value,
		})
		return
	}

	// append (resolve ns)
	n.Attributes = append(n.Attributes, &Attribute{
		Namespace: n.ResolveNS(namespace),
		Name:      name,
		Value:     value,
	})
}

// GetAttribute retrieves an attribute
func (n *Node) GetAttribute(name string) (*Attribute, bool) {
	for _, attr := range n.Attributes {
		if attr.Name == name {
			return attr, true
		}
	}
	return nil, false
}

// GetAttributeValue returns the value of a given attribute (empty string if it doesn't exist)
func (n *Node) GetAttributeValue(name string) string {
	if attr, exists := n.GetAttribute(name); exists {
		return attr.Value
	}
	return ""
}

// GetAttributeX retrieves an attribute (escaped name as input)
func (n *Node) GetAttributeX(name string) (*Attribute, bool) {
	for _, attr := range n.Attributes {
		if n.xPathEscape(attr.Name) == name {
			return attr, true
		}
	}
	return nil, false
}

// GetAttributeXValue returns the value of a given attribute (empty string if it doesn't exist) (escaped name as input)
func (n *Node) GetAttributeXValue(name string) string {
	if attr, exists := n.GetAttributeX(name); exists {
		return attr.Value
	}
	return ""
}

// ResolveNS tries to resolve the given Namespace
func (n *Node) ResolveNS(namespace string) *Namespace {
	return n.resolveNS(namespace, nil)
}

func (n *Node) resolveNS(namespace string, rel *Node) *Namespace {
	if namespace == "" {
		return nil
	}

	// see if we know this namespace
	if n.NamespaceMapping != nil {
		if ns, exists := n.NamespaceMapping[namespace]; exists {
			return ns
		}
	}

	// ask parent
	if n.Parent != nil {
		if parentResult := n.Parent.resolveNS(namespace, n); parentResult != nil {
			return parentResult
		}
	}

	// register new namespace
	if rel == nil {
		return n.RegisterNS(namespace, n.NewNSAbbrev())
	}
	return nil
}

// ResolveNSAbbrev tries to resolve the given Namespace by its Abbreviation
func (n *Node) ResolveNSAbbrev(abbreviation string) *Namespace {
	// see if we know this namespace
	if n.NamespaceMapping != nil {
		for _, chk := range n.NamespaceMapping {
			if chk.Abbreviation == abbreviation {
				return chk
			}
		}
	}

	// ask parent
	if n.Parent != nil {
		return n.Parent.ResolveNSAbbrev(abbreviation)
	}

	// bad luck
	panic(fmt.Errorf("ResolveNSAbbrev(): unknown namespace abbreviation: [%s]", abbreviation))
}

// SetDefaultNS sets the default namespace for this Node and all children
func (n *Node) SetDefaultNS(namespace string) {
	n.RegisterNS(namespace, "")
}

// RegisterNS registers the specified namespace as known for this node
func (n *Node) RegisterNS(namespace string, abbreviation string) *Namespace {
	if n.NamespaceMapping == nil {
		n.NamespaceMapping = map[string]*Namespace{}
	}

	newNamespace := &Namespace{
		Name:         namespace,
		Abbreviation: abbreviation,
	}

	n.NamespaceMapping[namespace] = newNamespace
	return newNamespace
}

// NewNSAbbrev returns a new abbreviation to use for a namespace
func (n *Node) NewNSAbbrev() string {
	if n.Document != nil {
		return n.Document.NewNSAbbrev()
	}
	panic(fmt.Errorf("NewNSAbbrev(): requires document, nil set"))
}

// CopyValue copies the source value to the given target Node
func (n *Node) CopyValue(target *Node) {
	n.Value.CopyValue(target)
}

// String returns the Value as string
func (n *Node) String() string {
	return n.Value.String()
}

// IsNil returns true if the stored value is nil
func (n *Node) IsNil() bool {
	return n.Value.IsNil()
}

// SetValue sets the value of this Node
func (n *Node) SetValue(val interface{}) {
	n.Value.SetValue(val)
}

// NewChildren appends a new child Node to this Node and returns it
func (n *Node) NewChildren(name string, namespace string) *Node {
	newNode := &Node{
		Name:     name,
		Exists:   true,
		Document: n.Document,
		Parent:   n,
	}
	n.Children.Append(newNode)
	newNode.Namespace = n.ResolveNS(namespace)
	return newNode
}

// ReplaceChildren replaces the children with the specified Node if it already exists, appends otherwise
func (n *Node) ReplaceChildren(name string, namespace string) *Node {
	for _, child := range n.Children.All() {
		if child.Name == name {
			child = &Node{
				Namespace: n.ResolveNS(namespace),
				Name:      name,
				Exists:    true,
				Document:  n.Document,
				Parent:    n,
			}
			return child
		}
	}
	return n.NewChildren(name, namespace)
}

func (n *Node) debug(indent int) {
	displayName := n.Name
	if n.Namespace != nil {
		displayName = n.Namespace.Abbreviation + ":" + displayName
	}

	fmt.Printf("%s- %s -> %s\n", strings.Repeat(" ", indent), displayName, n.String())

	if len(n.Attributes) > 0 {
		for _, attr := range n.Attributes {
			displayName := attr.Name
			if attr.Namespace != nil {
				displayName = attr.Namespace.Abbreviation + ":" + displayName
			}
			fmt.Printf("%s@%s -> %s\n", strings.Repeat(" ", indent+2), displayName, attr.Value)
		}
	}

	for _, child := range n.Children.All() {
		child.debug(indent + 2)
	}
}
