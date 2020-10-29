package wsdl

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/lordkhonsu/go-soap/dom"
)

// Request wraps a single SOAP request
type Request struct {
	client    *Client
	service   *Service
	wsdl      *WSDL
	operation *Operation

	envelope *dom.Document
	rootNode *dom.Node
	header   *dom.Node
	body     *dom.Node

	headerValues *dom.Document
	bodyValues   *dom.Document

	typeExtensions map[string]string
}

func (r *Request) init() {
	// build header+body values
	r.typeExtensions = map[string]string{}
	r.headerValues = dom.NewDocument("Header")
	r.bodyValues = dom.NewDocument("Body")

	// build envelope
	r.envelope = dom.NewDocument("s:Envelope")
	r.rootNode = r.envelope.Root
	r.rootNode.RegisterNS("http://www.w3.org/2001/XMLSchema-instance", "i")
	r.rootNode.RegisterNS("http://schemas.xmlsoap.org/soap/envelope/", "s")

	// attach header
	r.header = r.rootNode.NewChildren("Header", "http://schemas.xmlsoap.org/soap/envelope/")
	r.header.SetDefaultNS(r.client.targetNamespace)

	// attach body
	r.body = r.rootNode.NewChildren("Body", "http://schemas.xmlsoap.org/soap/envelope/")
	r.body.SetDefaultNS(r.client.targetNamespace)
}

// SetInputHeader sets the specified input header value
func (r *Request) SetInputHeader(name string, value interface{}) {
	node := r.headerValues.Root.ReplaceChildren(name, "")
	node.SetValue(value)
}

// SetInputHeaders sets the specified input header values
func (r *Request) SetInputHeaders(mapping map[string]interface{}) {
	for k, v := range mapping {
		r.SetInputHeader(k, v)
	}
}

// SetTypeExtension marks a type in the body to be replaced by the given extension
func (r *Request) SetTypeExtension(xpath string, fqType string) {
	r.typeExtensions[xpath] = fqType
}

// SetBodyValues sets the body values to use
func (r *Request) SetBodyValues(body *dom.Document) {
	r.bodyValues = body.Wrap("Body").Wrap("s:Envelope")
}

// XML returns the XML data for this Request
func (r *Request) XML() string {
	r.build()
	return r.envelope.XML()
}

// Send sends the request
func (r *Request) Send() *Response {
	body := bytes.NewBufferString(r.XML())

	httpRequest, err := http.NewRequest("POST", r.service.url, body)
	if err != nil {
		panic(err)
	}

	httpRequest.Header.Add("Content-Type", "text/xml; charset=utf-8")
	httpRequest.Header.Add("SOAPAction", r.GetSOAPAction())

	httpResponse, err := r.client.httpClient.Do(httpRequest)
	if err != nil {
		panic(err)
	}

	raw, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		panic(err)
	}

	return ParseResponse(raw)
}

func (r *Request) build() {
	r.buildHeader()
	r.buildBody()
}

// GetSOAPAction returns the named SOAP action this Request targets
func (r *Request) GetSOAPAction() string {
	return r.operation.domNode.XPath("operation").First().GetAttributeValue("soapAction")
}

func (r *Request) buildHeader() {
	// clear header
	r.header.Children.ClearAll()

	// write soap action
	soapAction := r.header.NewChildren("Action", "")
	soapAction.SetValue(r.GetSOAPAction())
	soapAction.SetAttribute("", "mustUnderstand", "1")

	// write header values
	inputHeaderList := r.operation.domNode.XPath("input/header")
	for _, inputHeader := range inputHeaderList.All() {
		name := inputHeader.GetAttributeValue("part")
		headerNode := r.header.NewChildren(name, "")
		if val := r.headerValues.XPath("%s", name).First(); val.Exists {
			val.CopyValue(headerNode)
		}
	}
}

func (r *Request) buildBody() {
	// clear body
	r.body.Children.ClearAll()

	// write body; find message for body
	_, messageName := dom.SplitFQName(r.operation.domNode.XPath("input").First().GetAttributeValue("message"))
	message := r.wsdl.XPath("/definitions/message[@name='%s']", messageName).First()
	messageElement := message.XPath("part").First().GetAttributeValue("element")

	// find element for message
	r.wsdl.FindElement(messageElement, message).Build(r.body, r.bodyValues, r.typeExtensions)
}
