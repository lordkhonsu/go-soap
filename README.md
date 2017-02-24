# go-soap

## Description

go-soap has the target to provide a simple access interface for working with SOAP endpoints specified by WSDL in Go

## Sample

```
package main

import (
	"fmt"
	"github.com/lordkhonsu/go-soap/dom"
	"github.com/lordkhonsu/go-soap/wsdl"
)

func main() {
	client := wsdl.NewClient("https://localhost:8080/Sample.svc?wsdl")
	request := client.Service("SampleService").Operation("SampleOperation").NewRequest()

	request.SetBodyValues(dom.Convert("SampleOperationMsg", map[string]interface{}{
		"Username":       "sample-user",
		"Password":       "sample-password,
	}))

	fmt.Println(request.XML())

	response := request.Send()

	if fault := response.Fault(); fault != nil {
		// a Fault occured
		panic(fault.Error())
	}

	fmt.Println(response.XML())

	// use response.Body().XPath(...) to fetch the values you need
}
```

## Derived types

You can register to use a derived type at an exact location by doing so:

```
	request.SetTypeExtension("/s:Envelope/Body[1]/SampleOperationMsg[1]", "tns:SampleDerivedOperationMsg")
```

## To-do

* Lots of XPath commands still missing
* Schema validation (enums, field requirements etc)
* CDATA for XML
* Tests

## License

MIT, see LICENSE.md