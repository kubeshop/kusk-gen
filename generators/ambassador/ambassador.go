package ambassador

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/kubeshop/kusk/generators"
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

func getServiceURL(options *generators.Options) string {
	if options.Service.Port > 0 {
		return fmt.Sprintf(
			"%s.%s:%d",
			options.Service.Name,
			options.Service.Namespace,
			options.Service.Port,
		)
	}

	return fmt.Sprintf("%s.%s", options.Service.Name, options.Service.Namespace)
}

func Generate(options *generators.Options, spec *openapi3.T) (string, error) {
	var mappings []mappingTemplateData

	serviceURL := getServiceURL(options)

	if options.SplitPaths {
		// generate a mapping for each operation

		for path, pathItem := range spec.Paths {
			for method, operation := range pathItem.Operations() {
				mappingPath, regex := generateMappingPath(path, operation)

				op := mappingTemplateData{
					MappingName:      generateMappingName(options.Service.Name, method, path, operation),
					MappingNamespace: options.Namespace,
					ServiceURL:       serviceURL,
					BasePath:         options.BasePath,
					TrimPrefix:       options.TrimPrefix,
					Method:           method,
					Path:             mappingPath,
					Regex:            regex,
				}

				mappings = append(mappings, op)
			}
		}
	} else {
		op := mappingTemplateData{
			MappingName:      options.Service.Name,
			MappingNamespace: options.Namespace,
			ServiceURL:       serviceURL,
			BasePath:         options.BasePath,
			TrimPrefix:       options.TrimPrefix,
		}

		mappings = append(mappings, op)
	}

	// We need to sort mappings as in the process of conversion of YAML to JSON
	// the Go map's access mechanics randomize the order and therefore the output is shuffled.
	// Not only it makes tests fail, it would also affect people who would use this in order to
	// generate manifests and use them in GitOps processes
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].MappingName < mappings[j].MappingName
	})

	var buf bytes.Buffer

	err := mappingTemplate.Execute(&buf, mappings)
	if err != nil {
		return "", err
	}

	res := buf.String()

	return reDuplicateNewlines.ReplaceAllString(res, "\n"), nil
}
