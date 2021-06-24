package ambassador

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	mappingTemplate *template.Template
)

type opTemplate struct {
	Namespace     string
	ServiceName   string
	OperationName string

	Method string
	Path   string
	Regex  bool
}

func init() {
	mappingTemplate = template.New("mapping")
	mappingTemplate = template.Must(mappingTemplate.Parse(mappingTemplateRaw))
}

// generateMappingPath returns the final pattern that should go to mapping
// and whether the regex should be used
func generateMappingPath(path string, op *openapi3.Operation) (string, bool) {
	containsPathParameter := false
	for _, param := range op.Parameters {
		if param.Value.In == "path" {
			containsPathParameter = true
			break
		}
	}

	if !containsPathParameter {
		return path, false
	}

	// replace each parameter with appropriate regex
	for _, param := range op.Parameters {
		if param.Value.In != "path" {
			continue
		}

		// the regex evaluation for mapping routes is actually done
		// within Envoy, which uses ECMA-262 regex grammar
		// https://www.envoyproxy.io/docs/envoy/v1.5.0/api-v1/route_config/route#route
		// https://en.cppreference.com/w/cpp/regex/ecmascript
		// https://www.getambassador.io/docs/edge-stack/latest/topics/using/rewrites/#regex_rewrite

		replaceWith := `([a-zA-Z0-9]*)`

		oldParam := "{" + param.Value.Name + "}"

		path = strings.ReplaceAll(path, oldParam, replaceWith)
	}

	return path, true
}

func GenerateMappings(namespace, serviceName string, spec *openapi3.T) (string, error) {
	var operations []opTemplate

	for path, pathItem := range spec.Paths {
		for method, operation := range pathItem.Operations() {
			mappingPath, regex := generateMappingPath(path, operation)

			op := opTemplate{
				Namespace:   namespace,
				ServiceName: serviceName,
				// TODO: OperationID may not be present, we'll have to generate something here
				OperationName: operation.OperationID,
				Method:        method,
				Path:          mappingPath,
				Regex:         regex,
			}

			operations = append(operations, op)
		}
	}

	var buf bytes.Buffer

	err := mappingTemplate.Execute(&buf, operations)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
