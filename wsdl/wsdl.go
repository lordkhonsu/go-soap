package wsdl

import (
	"fmt"
	"github.com/lordkhonsu/go-soap/dom"
)

// WSDL wraps a WSDL specific dom.Document and provides helper methods
type WSDL struct {
	document *dom.Document
}

// XPath resolves the given XPath to a list of Nodes
func (w *WSDL) XPath(xpath string, arguments ...interface{}) *dom.NodeList {
	return w.document.XPath(xpath, arguments...)
}

// FindElement finds the specified <element> in the WSDL and returns a dom.Node that represents it
func (w *WSDL) FindElement(fqName string, relative *dom.Node) *Element {
	elemNS, elemName := dom.SplitFQName(fqName)
	base := w.document.Root
	if relative != nil {
		base = relative
	}
	namespace := base.ResolveNSAbbrev(elemNS)

	elemNode := w.XPath("/definitions/types/schema[@targetNamespace='%s']/element[@name='%s']", namespace.Name, elemName).First()
	if !elemNode.Exists {
		panic(fmt.Errorf("FindElement(): element with fqName [%s] not found", fqName))
	}

	return &Element{
		wsdl:            w,
		targetNamespace: namespace.Name,
		domNode:         elemNode,
	}
}

// FindType finds the specified <type> in the WSDL and returns a dom.Node that represents it
func (w *WSDL) FindType(fqName string, relative *dom.Node) *Type {
	elemNS, elemName := dom.SplitFQName(fqName)
	base := w.document.Root
	if relative != nil {
		base = relative
	}
	namespace := base.ResolveNSAbbrev(elemNS)

	if namespace.Name == "http://www.w3.org/2001/XMLSchema" {
		return &Type{
			wsdl:            w,
			targetNamespace: namespace.Name,
			w3cType:         true,
			w3cName:         elemName,
		}
	}

	baseSchema := w.XPath("/definitions/types/schema[@targetNamespace='%s']", namespace.Name).First()

	// find complexType first
	elemNode := baseSchema.XPath("complexType[@name='%s']", elemName).First()

	// find simpleType next
	if !elemNode.Exists {
		elemNode = baseSchema.XPath("simpleType[@name='%s']", elemName).First()
	}

	// nothing found
	if !elemNode.Exists {
		panic(fmt.Errorf("FindType(): type with fqName [%s] not found in ns [%s]", fqName, namespace.Name))
	}

	return &Type{
		wsdl:            w,
		targetNamespace: namespace.Name,
		domNode:         elemNode,
	}
}
