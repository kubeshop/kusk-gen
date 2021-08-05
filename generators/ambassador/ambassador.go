package ambassador

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
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

func init() {
	generators.Registry["ambassador"] = &Generator{}
}

type Generator struct{}

func (g *Generator) ShortDescription() string {
	return "Generates Ambassador Mappings for your service"
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Cmd() string {
	return "ambassador"
}

func (g *Generator) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("ambassador", pflag.ExitOnError)

	fs.String(
		"path.base",
		"/",
		"a base path for Service endpoints",
	)

	fs.String(
		"path.trim_prefix",
		"",
		"a prefix to trim from the URL before forwarding to the upstream Service",
	)

	fs.Bool(
		"path.split",
		false,
		"force Kusk to generate a separate Mapping for each operation",
	)

	fs.Uint32(
		"timeouts.request_timeout",
		0,
		"total request timeout (seconds)",
	)

	fs.Uint32(
		"timeouts.idle_timeout",
		0,
		"idle connection timeout (seconds)",
	)

	return fs
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	if err := opts.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate options: %w", err)
	}

	var mappings []mappingTemplateData

	serviceURL := g.getServiceURL(opts)

	if g.shouldSplit(opts, spec) {
		// generate a mapping for each operation
		basePath := opts.Path.Base
		if basePath == "/" {
			basePath = ""
		}

		for path, pathItem := range spec.Paths {
			if pathSubOptions, ok := opts.PathSubOptions[path]; ok {
				if pathSubOptions.Disabled {
					continue
				}
			}

			for method, operation := range pathItem.Operations() {
				if opSubOptions, ok := opts.OperationSubOptions[method+path]; ok {
					if opSubOptions.Disabled {
						continue
					}
				}

				mappingPath, regex := g.generateMappingPath(path, operation)

				op := mappingTemplateData{
					MappingName:      g.generateMappingName(opts.Service.Name, method, path, operation),
					MappingNamespace: opts.Namespace,
					ServiceURL:       serviceURL,
					BasePath:         basePath,
					TrimPrefix:       opts.Path.TrimPrefix,
					Method:           method,
					Path:             mappingPath,
					Regex:            regex,
				}

				var corsOpts options.CORSOptions

				// take global CORS options
				corsOpts = opts.Ingress.CORS

				// if path-level CORS options are different, override with them
				if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
					if !reflect.DeepEqual(corsOpts, pathSubOpts.CORS) {
						corsOpts = pathSubOpts.CORS
					}
				}

				// if operation-level CORS options are different, override them
				if opSubOpts, ok := opts.OperationSubOptions[path]; ok {
					if !reflect.DeepEqual(corsOpts, opSubOpts.CORS) {
						corsOpts = opSubOpts.CORS
					}
				}

				// if final CORS options are not empty, include them
				if !reflect.DeepEqual(options.CORSOptions{}, corsOpts) {
					op.CORSEnabled = true
					op.CORS = g.corsTemplateData(&corsOpts)
				}

				var timeoutOpts options.TimeoutOptions

				// take global timeout options
				timeoutOpts = opts.Timeouts

				// if path-level timeout options are different, override with them
				if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
					if !reflect.DeepEqual(timeoutOpts, pathSubOpts.Timeouts) {
						timeoutOpts = pathSubOpts.Timeouts
					}
				}

				// if operation-level timeout options are different, override them
				if opSubOpts, ok := opts.OperationSubOptions[path]; ok {
					if !reflect.DeepEqual(timeoutOpts, opSubOpts.Timeouts) {
						timeoutOpts = opSubOpts.Timeouts
					}
				}

				// if final timeout options are not empty, include them
				if !reflect.DeepEqual(options.TimeoutOptions{}, timeoutOpts) {
					op.RequestTimeout = timeoutOpts.RequestTimeout * 1000
					op.IdleTimeout = timeoutOpts.IdleTimeout * 1000
				}

				mappings = append(mappings, op)
			}
		}
	} else {
		op := mappingTemplateData{
			MappingName:      opts.Service.Name,
			MappingNamespace: opts.Namespace,
			ServiceURL:       serviceURL,
			BasePath:         opts.Path.Base,
			TrimPrefix:       opts.Path.TrimPrefix,
			RequestTimeout:   opts.Timeouts.RequestTimeout * 1000,
			IdleTimeout:      opts.Timeouts.IdleTimeout * 1000,
		}

		// if global CORS options are defined, take them
		if !reflect.DeepEqual(options.CORSOptions{}, opts.Ingress.CORS) {
			op.CORSEnabled = true
			op.CORS = g.corsTemplateData(&opts.Ingress.CORS)
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

// generateMappingPath returns the final pattern that should go to mapping
// and whether the regex should be used
func (g *Generator) generateMappingPath(path string, op *openapi3.Operation) (string, bool) {
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

func (g *Generator) generateMappingName(serviceName, method, path string, operation *openapi3.Operation) string {
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

func (g *Generator) getServiceURL(options *options.Options) string {
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

func (g *Generator) shouldSplit(opts *options.Options, spec *openapi3.T) bool {
	if opts.Path.Split {
		return true
	}

	for path, pathItem := range spec.Paths {
		if pathSubOptions, ok := opts.PathSubOptions[path]; ok {
			// a path is disabled
			if pathSubOptions.Disabled {
				return true
			}

			// a path has non-zero, different from global scope CORS options
			if !reflect.DeepEqual(options.CORSOptions{}, pathSubOptions.CORS) &&
				!reflect.DeepEqual(opts.Ingress.CORS, pathSubOptions.CORS) {
				return true
			}

			// a path has non-zero, different from global scope timeouts options
			if !reflect.DeepEqual(options.TimeoutOptions{}, pathSubOptions.Timeouts) &&
				!reflect.DeepEqual(opts.Timeouts, pathSubOptions.Timeouts) {
				return true
			}
		}

		for method := range pathItem.Operations() {
			if opSubOptions, ok := opts.OperationSubOptions[method+path]; ok {
				// an operation is disabled
				if opSubOptions.Disabled {
					return true
				}

				// an operation has non-zero, different from global CORS options
				if !reflect.DeepEqual(options.CORSOptions{}, opSubOptions.CORS) &&
					!reflect.DeepEqual(opts.Ingress.CORS, opSubOptions.CORS) {
					return true
				}

				// an operation has non-zero, different from global timeouts options
				if !reflect.DeepEqual(options.TimeoutOptions{}, opSubOptions.Timeouts) &&
					!reflect.DeepEqual(opts.Timeouts, opSubOptions.Timeouts) {
					return true
				}
			}
		}
	}

	return false
}

func (g *Generator) corsTemplateData(corsOpts *options.CORSOptions) corsTemplateData {
	res := corsTemplateData{
		Origins:        strings.Join(corsOpts.Origins, ","),
		Methods:        strings.Join(corsOpts.Methods, ","),
		Headers:        strings.Join(corsOpts.Headers, ","),
		ExposedHeaders: strings.Join(corsOpts.ExposeHeaders, ","),
		MaxAge:         fmt.Sprint(corsOpts.MaxAge),
	}

	if corsOpts.Credentials != nil {
		res.Credentials = *corsOpts.Credentials
	}

	return res
}
