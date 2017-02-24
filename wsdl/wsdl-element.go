package wsdl

import (
	"fmt"
	"github.com/lordkhonsu/go-soap/dom"
	"strconv"
)

// Element wraps the WSDL specification for an <element> and provides helper methods
type Element struct {
	wsdl            *WSDL
	targetNamespace string
	domNode         *dom.Node
}

func (e *Element) resolveType() *Type {
	if typeName := e.domNode.GetAttributeValue("type"); len(typeName) > 0 {
		return e.wsdl.FindType(typeName, e.domNode)

	} else if embedded := e.domNode.XPath("complexType").First(); embedded.Exists {
		return &Type{
			wsdl:            e.wsdl,
			targetNamespace: e.targetNamespace,
			domNode:         embedded,
		}

	} else if embedded := e.domNode.XPath("simpleType").First(); embedded.Exists {
		return &Type{
			wsdl:            e.wsdl,
			targetNamespace: e.targetNamespace,
			domNode:         embedded,
		}

	}

	panic(fmt.Errorf("resolveType(): type for element [%s] could not be resolved", e.Name()))
}

// Name returns the Name for this <element>
func (e *Element) Name() string {
	return e.domNode.GetAttributeValue("name")
}

// MinOccurs returns the minimum amount this element must appear
func (e *Element) MinOccurs() int {
	if v := e.domNode.GetAttributeValue("minOccurs"); len(v) > 0 {
		buffer, _ := strconv.ParseInt(v, 10, 64)
		return int(buffer)
	}
	return 1
}

// MaxOccurs returns the maximum amount this element may appear (-1 = unlimited)
func (e *Element) MaxOccurs() int {
	if v := e.domNode.GetAttributeValue("maxOccurs"); len(v) > 0 {
		if v == "unbounded" {
			return -1
		}
		buffer, _ := strconv.ParseInt(v, 10, 64)
		return int(buffer)
	}
	return 1
}

// IsNillable returns true if this <element> may be nil
func (e *Element) IsNillable() bool {
	if v := e.domNode.GetAttributeValue("nillable"); len(v) > 0 {
		return v == "true"
	}
	return false
}

// Build builds this <element> and attaches it to the given parent dom.Node
func (e *Element) Build(parent *dom.Node, body *dom.Document, typeExtensions map[string]string) {
	myType := e.resolveType()

	myXPath := parent.GetXPath() + "/" + parent.GetXPathName(e.Name())
	if extension, exists := typeExtensions[myXPath]; exists {
		myType = e.wsdl.FindType(extension, e.domNode)
	}

	// count values
	values := body.XPath(parent.GetXPath() + "/" + e.Name()).Len()
	count := 0
	minOccurs := e.MinOccurs()
	maxOccurs := e.MaxOccurs()

	/*
		fmt.Printf("name = %s, min = %d, max = %d, values = %d, nillable = %v\n",
			e.Name(), minOccurs, maxOccurs, values, e.IsNillable())
	*/

	if minOccurs == 0 && values == 0 {
		// skip if we may
		return
	}

	for {
		myType.build(parent, e.Name(), e.targetNamespace, body, typeExtensions)
		count++
		if (maxOccurs >= 0 && count >= maxOccurs) || (count >= minOccurs && values <= count) {
			// we have reached our end, stop here
			break
		}
	}

	missing := minOccurs - count
	if missing > 0 {
		for i := 0; i < missing; i++ {
			// we are missing some, fill
			myType.build(parent, e.Name(), e.targetNamespace, body, typeExtensions)
		}
	}
}
