package spec

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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

// Parse is the entrypoint for the spec package
// Accepts a path that should be parseable into a resource locater
// i.e. a URL or relative file path
func Parse(path string) (*openapi3.T, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("could not parse resource path: %w", err)
	}

	if isURLRelative := u.Host == ""; isURLRelative {
		return parseFromFile(path)
	}

	return parseFromURL(u)
}

func parseFromURL(u *url.URL) (*openapi3.T, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to read spec from url %s: %w", u.String(), err)
	}

	defer resp.Body.Close()

	return ParseFromReader(resp.Body)
}

func parseFromFile(path string) (*openapi3.T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open api spec file: %w", err)
	}

	defer f.Close()

	return ParseFromReader(f)
}

// ParseFromReader allows for providing your own Reader implementation
// to parse the API spec from
func ParseFromReader(contents io.Reader) (*openapi3.T, error) {
	spec, err := ioutil.ReadAll(contents)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of api spec: %w", err)
	}

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
