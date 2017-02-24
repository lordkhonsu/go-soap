package dom

import (
	"encoding/xml"
	"fmt"
)

// Document defines the topmost DOM entry
type Document struct {
	Root         *Node
	nextNSAbbrev int
}

// NewDocument returns a new document ready to use
func NewDocument(rootNode string) *Document {
	newDocument := &Document{Root: &Node{
		Name:   rootNode,
		Exists: true,
	}}
	newDocument.Root.Document = newDocument
	return newDocument
}

// Convert takes the argument and converts it into our DOM
func Convert(name string, in interface{}) *Document {
	result := NewDocument(name)
	result.Root.convert(name, in, result, nil)
	return result
}

func (n *Node) convert(name string, in interface{}, document *Document, parent *Node) {
	n.Name = name
	n.Document = document
	n.Parent = parent

	switch t := in.(type) {
	case map[string]interface{}:
		for k, v := range t {
			n.NewChildren(k, "").convert(k, v, document, n)
		}
	case []map[string]interface{}:
		for _, sub := range t {
			for k, v := range sub {
				n.NewChildren(k, "").convert(k, v, document, n)
			}
		}
	case Map:
		for k, v := range t {
			n.NewChildren(k, "").convert(k, v, document, n)
		}
	case []Map:
		for _, sub := range t {
			for k, v := range sub {
				n.NewChildren(k, "").convert(k, v, document, n)
			}
		}
	default:
		n.Value.value = t
	}
}

// XML outputs the Node as a XML entity
func (d *Document) XML() string {
	return xml.Header + d.Root.XML()
}

// UnmarshalXML implements the xml.Unmarshaler interface for decoding
func (d *Document) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	if d.Root == nil {
		d.Root = &Node{Document: d}
	}
	return dec.DecodeElement(d.Root, &start)
}

// XPath resolves the given XPath to a list of Nodes
func (d *Document) XPath(xpath string, arguments ...interface{}) *NodeList {
	return d.Root.XPath(xpath, arguments...)
}

// NewNSAbbrev returns a new abbreviation to use for a namespace
func (d *Document) NewNSAbbrev() string {
	d.nextNSAbbrev++
	return fmt.Sprintf("q%d", d.nextNSAbbrev)
}

// Debug outputs the entire Document
func (d *Document) Debug() {
	d.Root.debug(0)
}

// Wrap wraps the Root node in a new Root node with the given element name
func (d *Document) Wrap(name string) *Document {
	oldRoot := d.Root

	d.Root = &Node{
		Name:     name,
		Document: d,
		Exists:   true,
	}

	oldRoot.Parent = d.Root
	d.Root.Children.Append(oldRoot)

	return d
}
