package wsdl

import (
	"github.com/lordkhonsu/go-soap/dom"
)

// Type wraps the WSDL specification for an <complexType> and <simpleType> and provides helper methods
type Type struct {
	wsdl            *WSDL
	targetNamespace string
	domNode         *dom.Node
	w3cType         bool
	w3cName         string
}

func (t *Type) build(parent *dom.Node, name string, namespace string, body *dom.Document, typeExtensions map[string]string) *dom.Node {
	var self *dom.Node

	// basic type, nothing to do here
	if t.w3cType {
		self = parent.NewChildren(name, namespace)

	} else {
		// complex content -> base class
		xPathPrefix := ""
		if extension := t.domNode.XPath("complexContent/extension").First(); extension.Exists {
			xPathPrefix = "complexContent/extension/"
			baseType := t.wsdl.FindType(extension.GetAttributeValue("base"), t.domNode)
			self = baseType.build(parent, name, namespace, body, typeExtensions)
			self.SetAttribute("http://www.w3.org/2001/XMLSchema-instance", "type", t.domNode.GetAttributeValue("name"))
		} else {
			self = parent.NewChildren(name, namespace)
		}

		// embedded complex type -> sequence of new elements
		if sequence := t.domNode.XPath(xPathPrefix + "sequence/element"); sequence.Len() > 0 {
			for _, domNode := range sequence.All() {
				element := &Element{
					wsdl:            t.wsdl,
					targetNamespace: t.targetNamespace,
					domNode:         domNode,
				}

				element.Build(self, body, typeExtensions)
			}
		}
	}

	if val := body.XPath(self.GetXPath()).First(); val.Exists && !val.IsNil() {
		val.CopyValue(self)
	}

	if self.Children.Len() == 0 {
		if self.IsNil() {
			self.SetAttribute("http://www.w3.org/2001/XMLSchema-instance", "nil", "true")
		} else {
			self.SetAttribute("http://www.w3.org/2001/XMLSchema-instance", "nil", "false")
		}
	}

	return self
}

func (t *Type) debug() string {
	if t.w3cType {
		return "w3cType - " + t.w3cName
	}
	return t.domNode.Name
}
