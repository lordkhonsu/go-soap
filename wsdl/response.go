package wsdl

import (
	"encoding/xml"
	"github.com/lordkhonsu/go-soap/dom"
)

// Response wraps the received SOAP response
type Response struct {
	document *dom.Document
}

// ParseResponse parses the XML passed in raw and returns a Response
func ParseResponse(raw []byte) *Response {
	newResponse := &Response{
		document: dom.NewDocument("response"),
	}
	xml.Unmarshal(raw, newResponse.document)
	return newResponse
}

// XML returns the Response encoded as XML
func (r *Response) XML() string {
	return r.document.XML()
}

// XPath resolves the given XPath to a list of Nodes
func (r *Response) XPath(xpath string, arguments ...interface{}) *dom.NodeList {
	return r.document.XPath(xpath, arguments...)
}

// Header returns the contained Header
func (r *Response) Header() *dom.Node {
	return r.XPath("/Envelope/Header").First()
}

// Body returns the contained Body
func (r *Response) Body() *dom.Node {
	return r.XPath("/Envelope/Body").First()
}

// Fault returns a Fault struct if the response encapsulates a Fault, nil otherwise
func (r *Response) Fault() *Fault {
	if faultDOM := r.Body().XPath("Fault").First(); faultDOM.Exists {
		return &Fault{
			response: r,
			domNode:  faultDOM,
		}
	}
	return nil
}
