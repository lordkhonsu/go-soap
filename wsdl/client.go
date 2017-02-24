package wsdl

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/lordkhonsu/go-soap/dom"
	"io/ioutil"
	"net/http"
)

// Client defines a WSDL parsing client, providing SOAP functions and entities
type Client struct {
	url             string
	httpClient      *http.Client
	wsdl            *WSDL
	name            string
	targetNamespace string
	services        map[string]*Service
}

// NewClient creates a new client given a WSDL specification at [url]
func NewClient(url string) *Client {
	client := &Client{
		url:        url,
		httpClient: &http.Client{},
	}
	client.init()
	return client
}

func (c *Client) init() {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		panic(err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	// decode WSDL XML as DOM Document
	c.wsdl = &WSDL{document: &dom.Document{}}
	decoder := xml.NewDecoder(bytes.NewBuffer(data))
	decoder.Strict = true
	decoder.Decode(&c.wsdl.document)

	c.name = c.wsdl.document.Root.GetAttributeValue("name")
	c.targetNamespace = c.wsdl.document.Root.GetAttributeValue("targetNamespace")

	// build services
	c.services = map[string]*Service{}
	list := c.wsdl.XPath("/definitions/service")
	for _, item := range list.All() {
		name := item.GetAttributeValue("name")
		c.services[name] = &Service{
			client: c,
			wsdl:   c.wsdl,
			name:   name,
			url:    item.XPath("port/address").First().GetAttributeValue("location"),
		}
		c.services[name].init()
	}
}

// Explain outputs all available Services
func (c *Client) Explain() {
	fmt.Printf("[ %s :: Services ]\n", c.name)
	for name, service := range c.services {
		fmt.Printf("- %s -> %s\n", name, service.url)
		service.explain(2)
	}
}

// Service returns the named Service
func (c *Client) Service(name string) *Service {
	if service, exists := c.services[name]; exists {
		return service
	}
	panic(fmt.Errorf("Service(): unknown service with name [%s]", name))
}
