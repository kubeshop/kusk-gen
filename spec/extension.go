package spec

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	"github.com/kubeshop/kusk/options"
)

// GetOptions would retrieve and parse x-kusk top-level OpenAPI extension
// that contains Kusk options. If there's no extension found, an empty object will be returned.
func GetOptions(spec *openapi3.T) (*options.Options, error) {
	var res options.Options

	if extension, ok := spec.Extensions["x-kusk"]; ok {
		if kuskExtension, ok := extension.(json.RawMessage); ok {
			err := yaml.Unmarshal(kuskExtension, &res)
			if err != nil {
				return nil, fmt.Errorf("failed to parse extension: %w", err)
			}
		}
	}

	return &res, nil
}
