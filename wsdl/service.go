package wsdl

import (
	"fmt"
	"strings"
)

// Service defines a single service endpoint in a WSDL client
type Service struct {
	client     *Client
	wsdl       *WSDL
	name       string
	url        string
	operations map[string]*Operation
}

func (s *Service) init() {
	// map all operations via binding ports
	s.operations = map[string]*Operation{}
	ports := s.client.wsdl.XPath("/definitions/service[@name='%s']/port", s.name)
	for _, port := range ports.All() {
		operations := s.client.wsdl.XPath("/definitions/binding[@name='%s']/operation", port.GetAttributeValue("name"))
		for _, operation := range operations.All() {
			name := operation.GetAttributeValue("name")
			s.operations[name] = &Operation{
				client:  s.client,
				service: s,
				wsdl:    s.wsdl,
				name:    name,
				domNode: operation,
			}
		}
	}
}

// Explain outputs all available operations
func (s *Service) Explain() {
	s.explain(0)
}

func (s *Service) explain(indent int) {
	indentation := strings.Repeat(" ", indent)
	fmt.Printf("%s[ %s :: Operations ]\n", indentation, s.name)

	for name := range s.operations {
		fmt.Printf("%s- %s\n", indentation, name)
	}
}

// Operation returns the named Operation
func (s *Service) Operation(name string) *Operation {
	if operation, exists := s.operations[name]; exists {
		return operation
	}
	panic(fmt.Errorf("Operation(): unknown operation with name [%s]", name))
}
