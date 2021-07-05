package spec

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

type header struct {
	Swagger string `json:"swagger"`
	OpenAPI string `json:"openapi"` // we might need that later to distinguish 3.1.x vs 3.0.x
}

// isSwagger tries to decode the spec header
func isSwagger(spec []byte) bool {
	var header header

	_ = yaml.Unmarshal(spec, &header)

	return header.Swagger != ""
}

func ParseFromFile(path string) (*openapi3.T, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	return Parse(contents)
}

func Parse(spec []byte) (*openapi3.T, error) {
	if isSwagger(spec) {
		return parseSwagger(spec)
	}

	return parseOpenAPI3(spec)
}

func parseSwagger(spec []byte) (*openapi3.T, error) {
	spec, err := yaml.YAMLToJSON(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	var swaggerSpec openapi2.T

	err = swaggerSpec.UnmarshalJSON(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Swagger: %w", err)
	}

	return openapi2conv.ToV3(&swaggerSpec)
}

func parseOpenAPI3(spec []byte) (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromData(spec)
}
