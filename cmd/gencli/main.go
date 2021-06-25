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

	mappings, err := ambassador.GenerateMappings(ambassador.Options{
		ServiceNamespace: "default",
		ServiceName:      "petstore",
		BasePath:         "/petstore/api/v3",
		TrimPrefix:       "/petstore",
		RootOnly:         true,
	}, spec)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mappings)
}
