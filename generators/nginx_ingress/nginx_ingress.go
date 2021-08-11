package nginx_ingress

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
)

const (
	ingressAPIVersion = "networking.k8s.io/v1"
	ingressKind       = "Ingress"
)

var (
	ingressClassName = "nginx"
	pathTypePrefix   = v1.PathTypePrefix
	pathTypeExact    = v1.PathTypeExact

	openApiPathVariableRegex = regexp.MustCompile(`{[A-z]+}`)
)

func init() {
	generators.Registry["nginx-ingress"] = &Generator{}
}

type Generator struct{}

func (g *Generator) Cmd() string {
	return "nginx-ingress"
}

func (g *Generator) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("nginx-ingress", pflag.ExitOnError)

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
		"force Kusk to generate a separate Ingress for each operation",
	)

	fs.String(
		"ingress.host",
		"",
		"an Ingress Host to listen on",
	)

	fs.Uint32(
		"timeouts.request_timeout",
		0,
		"total request timeout (seconds)",
	)

	fs.String(
		"nginx_ingress.rewrite_target",
		"",
		"a custom NGINX rewrite target",
	)

	return fs
}

func (g *Generator) ShortDescription() string {
	return "Generates nginx-ingress resources"
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	if err := opts.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate opts: %w", err)
	}

	ingresses := make([]v1.Ingress, 0)

	if g.shouldSplit(opts, spec) {
		for path := range spec.Paths {
			if pathSubOptions, ok := opts.PathSubOptions[path]; ok {
				if pathSubOptions.Disabled {
					continue
				}
			}

			name := fmt.Sprintf("%s-%s", opts.Service.Name, ingressResourceNameFromPath(path))

			var corsOpts options.CORSOptions

			// take global CORS options
			corsOpts = opts.CORS

			// if path-level CORS options are different, override with them
			if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
				if !reflect.DeepEqual(corsOpts, pathSubOpts.CORS) {
					corsOpts = pathSubOpts.CORS
				}
			}

			var timeoutOpts options.TimeoutOptions

			// take global Timeout options
			timeoutOpts = opts.Timeouts

			// if path-level CORS options are different, override with them
			if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
				if !reflect.DeepEqual(timeoutOpts, pathSubOpts.Timeouts) {
					timeoutOpts = pathSubOpts.Timeouts
				}
			}

			// Get initial set of annotation based on current options
			// will be modified next based on current path
			annotations := g.generateAnnotations(
				&opts.Path,
				&opts.NGINXIngress,
				&corsOpts,
				&timeoutOpts,
			)

			// if path has a parameter, replace {param} with ([A-z0-9]+) and set use regex annotation to true
			// if path has no parameter, just use path
			pathField := g.generatePath(&opts.Path, &opts.NGINXIngress)
			if openApiPathVariableRegex.MatchString(path) {
				pathField = opts.Path.Base + string(openApiPathVariableRegex.ReplaceAll([]byte(path), []byte("([A-z0-9]+)")))

				// get the first capture group of regex. Given a path /books/{id}, will return /books/
				rewrite := string(openApiPathVariableRegex.ReplaceAllLiteral([]byte(path), []byte("$1")))
				annotations[rewriteTargetAnnotationKey] = rewrite
				annotations[useRegexAnnotationKey] = "true"
			} else if path == "/" {
				pathField = opts.Path.Base + "$"
				annotations[rewriteTargetAnnotationKey] = "/"
				annotations[useRegexAnnotationKey] = "true"
			} else {
				pathField = opts.Path.Base + path
				annotations[rewriteTargetAnnotationKey] = path
			}

			// Replace // with /
			pathField = strings.ReplaceAll(pathField, "//", "/")

			ingress := g.newIngressResource(
				name,
				opts.Namespace,
				pathField,
				pathTypeExact,
				annotations,
				&opts.Service,
				opts.Host,
			)

			ingresses = append(ingresses, ingress)
		}
	} else {
		ingress := g.newIngressResource(
			fmt.Sprintf("%s-ingress", opts.Service.Name),
			opts.Namespace,
			g.generatePath(&opts.Path, &opts.NGINXIngress),
			pathTypePrefix,
			g.generateAnnotations(&opts.Path, &opts.NGINXIngress, &opts.CORS, &opts.Timeouts),
			&opts.Service,
			opts.Host,
		)
		ingresses = append(ingresses, ingress)
	}

	// We need to sort the ingresses as in the process of conversion of YAML to JSON
	// the Go map's access mechanics randomize the order and therefore the output is shuffled.
	// Not only it makes tests fail, it would also affect people who would use this in order to
	// generate manifests and use them in GitOps processes
	sort.Slice(ingresses, func(i, j int) bool {
		return ingresses[i].Name < ingresses[j].Name
	})

	return buildOutput(ingresses)
}

// Build suitable output to be piped into kubectl or a file
func buildOutput(ingresses []v1.Ingress) (string, error) {
	var builder strings.Builder

	for _, ingress := range ingresses {
		builder.WriteString("---\n") // indicate start of YAML resource
		b, err := yaml.Marshal(ingress)
		if err != nil {
			return "", fmt.Errorf("unable to marshal ingress resource: %+v: %s", ingress, err.Error())
		}
		builder.WriteString(string(b))
	}

	return builder.String(), nil
}

// Given a path such as /books/{id} return a suitable ingress resource name
// in the form books-id or root if the path is simply /
func ingressResourceNameFromPath(path string) string {
	if len(path) == 0 || path == "/" {
		return "root"
	}

	var b strings.Builder
	for _, pathItem := range strings.Split(path, "/") {
		if pathItem == "" {
			continue
		}

		// remove openapi path variable curly braces from path item
		strippedPathItem := strings.ReplaceAll(strings.ReplaceAll(pathItem, "{", ""), "}", "")
		fmt.Fprintf(&b, "%s-", strippedPathItem)
	}

	// remove trailing - character
	return strings.TrimSuffix(b.String(), "-")
}

func (g *Generator) newIngressResource(
	name,
	namespace,
	path string,
	pathType v1.PathType,
	annotations map[string]string,
	serviceOpts *options.ServiceOptions,
	host string,
) v1.Ingress {
	return v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ingressAPIVersion,
			Kind:       ingressKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: v1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []v1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									PathType: &pathType,
									Path:     path,
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: serviceOpts.Name,
											Port: v1.ServiceBackendPort{
												Number: serviceOpts.Port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
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
				!reflect.DeepEqual(opts.CORS, pathSubOptions.CORS) {
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
					!reflect.DeepEqual(opts.CORS, opSubOptions.CORS) {
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

func (g *Generator) generatePath(path *options.PathOptions, nginx *options.NGINXIngressOptions) string {
	if len(path.TrimPrefix) > 0 &&
		strings.HasPrefix(path.Base, path.TrimPrefix) &&
		nginx.RewriteTarget == "" {
		pathSuffixRegex := "(/|$)(.*)"

		return path.Base + pathSuffixRegex
	}

	return path.Base
}
