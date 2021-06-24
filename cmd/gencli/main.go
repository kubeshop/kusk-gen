package main

import (
	"context"
	"fmt"
	"log"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/kubeshop/openapi-kgen/ambassador"
)

func main() {
	loader := openapi3.Loader{
		Context: context.Background(),
	}

	spec, err := loader.LoadFromFile("petstore.yaml")
	if err != nil {
		log.Fatal(err)
	}

	mappings, err := ambassador.GenerateMappings("default", "svc", spec)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mappings)
}
