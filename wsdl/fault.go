package wsdl

import (
	"fmt"
	"github.com/lordkhonsu/go-soap/dom"
)

// Fault wraps a WSDL fault message
type Fault struct {
	response *Response
	domNode  *dom.Node
}

// String returns a formatted Fault message
func (f *Fault) String() string {
	return "[wsdl/fault] " + f.domNode.XPath("faultcode").First().String() + " - " +
		f.domNode.XPath("faultstring").First().String()
}

// Error returns the Fault as an error object
func (f *Fault) Error() error {
	return fmt.Errorf(f.String())
}

// Details returns all embedded fault detail structures (if there are any)
func (f *Fault) Details() *dom.NodeList {
	return f.domNode.XPath("detail/*")
}
