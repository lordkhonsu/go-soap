package wsdl

import (
	"github.com/lordkhonsu/go-soap/dom"
)

// Operation defines a single service operation
type Operation struct {
	client  *Client
	wsdl    *WSDL
	service *Service
	name    string
	domNode *dom.Node
}

// NewRequest creates a new Request instance for this Operation
func (o *Operation) NewRequest() *Request {
	newRequest := &Request{
		client:    o.client,
		service:   o.service,
		wsdl:      o.wsdl,
		operation: o,
	}
	newRequest.init()
	return newRequest
}
