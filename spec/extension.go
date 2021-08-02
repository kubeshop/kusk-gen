package spec

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	"github.com/kubeshop/kusk/options"
)

const kuskExtensionKey = "x-kusk"

func GetPathOptions(path *openapi3.PathItem) (*options.SubOptions, error) {
	var res options.SubOptions

	err := parseExtension(&path.ExtensionProps, &res)

	return &res, err
}

func GetOperationOptions(operation *openapi3.Operation) (*options.SubOptions, error) {
	var res options.SubOptions

	err := parseExtension(&operation.ExtensionProps, &res)

	return &res, err
}

// GetOptions would retrieve and parse x-kusk top-level OpenAPI extension
// that contains Kusk options. If there's no extension found, an empty object will be returned.
func GetOptions(spec *openapi3.T) (*options.Options, error) {
	var res options.Options

	if err := parseExtension(&spec.ExtensionProps, &res); err != nil {
		return nil, err
	}

	for pathString, path := range spec.Paths {
		var pathOpts options.Options

		if extension, ok := path.Extensions[kuskExtensionKey]; ok {
			if kuskExtension, ok := extension.(json.RawMessage); ok {
				err := yaml.Unmarshal(kuskExtension, &pathOpts)
				if err != nil {
					return nil, fmt.Errorf("failed to parse extension: %w", err)
				}

				if res.PathOperations == nil {
					res.PathOperations = map[string]options.Options{}
				}

				res.PathOperations[pathString] = pathOpts
			}
		}

		for method, operation := range path.Operations() {
			if extension, ok := operation.Extensions[kuskExtensionKey]; ok {
				if kuskExtension, ok := extension.(json.RawMessage); ok {
					var methodOpts options.Options
					err := yaml.Unmarshal(kuskExtension, &methodOpts)
					if err != nil {
						return nil, fmt.Errorf("failed to parse extension: %w", err)
					}

					if res.PathOperations == nil {
						res.PathOperations = map[string]options.Options{}
					}

					pathOpts = res.PathOperations[pathString]

					if pathOpts.HTTPMethodOperations == nil {
						pathOpts.HTTPMethodOperations = map[string]options.Options{}
					}

					pathOpts.HTTPMethodOperations[method] = methodOpts
					res.PathOperations[pathString] = pathOpts
				}
			}
		}
	}

	return &res, nil
}

func parseExtension(extensionProps *openapi3.ExtensionProps, target interface{}) error {
	if extension, ok := extensionProps.Extensions[kuskExtensionKey]; ok {
		if kuskExtension, ok := extension.(json.RawMessage); ok {
			err := yaml.Unmarshal(kuskExtension, target)
			if err != nil {
				return fmt.Errorf("failed to parse extension: %w", err)
			}
		}
	}

	return nil
}
