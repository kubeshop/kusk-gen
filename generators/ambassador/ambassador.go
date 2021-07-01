package ambassador

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	mappingTemplate     *template.Template
	reDuplicateNewlines = regexp.MustCompile(`\s*\n+`)
	rePathSymbols       = regexp.MustCompile(`[/{}]`)
)

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

func generateMappingName(serviceName, method, path string, operation *openapi3.Operation) string {
	var res strings.Builder

	if operation.OperationID != "" {
		res.WriteString(serviceName)
		res.WriteString("-")
		res.WriteString(operation.OperationID)
		return strings.ToLower(res.String())
	}

	// generate proper mapping name if operationId is missing
	res.WriteString(serviceName)
	res.WriteString("-")
	res.WriteString(method)
	res.WriteString(rePathSymbols.ReplaceAllString(path, ""))

	return strings.ToLower(res.String())
}

func GenerateMappings(options Options, spec *openapi3.T) (string, error) {
	if options.AmbassadorNamespace == "" {
		options.AmbassadorNamespace = "ambassador"
	}

	var mappings []mappingTemplateData

	if options.RootOnly && options.BasePath != "" {
		// generate a single mapping for the service
		op := mappingTemplateData{
			MappingName:         options.ServiceName,
			AmbassadorNamespace: options.AmbassadorNamespace,
			ServiceNamespace:    options.ServiceNamespace,
			ServiceName:         options.ServiceName,
			BasePath:            options.BasePath,
			TrimPrefix:          options.TrimPrefix,
		}

		mappings = append(mappings, op)
	} else {
		// generate a mapping for each operation

		for path, pathItem := range spec.Paths {
			for method, operation := range pathItem.Operations() {
				mappingPath, regex := generateMappingPath(path, operation)

				op := mappingTemplateData{
					MappingName:         generateMappingName(options.ServiceName, method, path, operation),
					AmbassadorNamespace: options.AmbassadorNamespace,
					ServiceNamespace:    options.ServiceNamespace,
					ServiceName:         options.ServiceName,
					BasePath:            options.BasePath,
					TrimPrefix:          options.TrimPrefix,
					Method:              method,
					Path:                mappingPath,
					Regex:               regex,
				}

				mappings = append(mappings, op)
			}
		}
	}

	var buf bytes.Buffer

	err := mappingTemplate.Execute(&buf, mappings)
	if err != nil {
		return "", err
	}

	res := buf.String()

	return reDuplicateNewlines.ReplaceAllString(res, "\n"), nil
}
