package nginx_ingress

import (
	"fmt"
	"reflect"
	"regexp"
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

	fs.String(
		"ingress.host",
		"",
		"an Ingress Host to listen on",
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

	pathsGenerated := map[string]struct{}{}

	for path, subOpts := range opts.PathSubOptions {
		if g.shouldSplit(opts, &subOpts) {
			// Mark the path as having a resource generated for it
			pathsGenerated[path] = struct{}{}

			name := ingressResourceNameFromPath(path)

			ingressOpts := options.IngressOptions{
				Host: opts.Ingress.Host,
				CORS: subOpts.CORS,
			}

			// if path has a parameter, replace {param} with ([A-z0-9]+) and set use regex annotation to true
			// if path has no parameter, just use path
			pathField := path
			annotations := g.generateAnnotations(&opts.Path, &opts.NGINXIngress, &ingressOpts.CORS)
			if openApiPathVariableRegex.MatchString(path) {
				pathField = string(openApiPathVariableRegex.ReplaceAll([]byte(path), []byte("([A-z0-9]+)")))

				// get the first capture group of regex. Given a path /books/{id}, will return /books/
				rewrite := string(openApiPathVariableRegex.ReplaceAllLiteral([]byte(path), []byte("$1")))
				annotations[rewriteTargetAnnotationKey] = rewrite
				annotations["nginx.ingress.kubernetes.io/use-regex"] = "true"
			}

			ingress := g.newIngressResource(
				name,
				opts.Namespace,
				opts.Path.Base + pathField,
				pathTypeExact,
				annotations,
				&opts.Service,
				&ingressOpts,
			)

			ingresses = append(ingresses, ingress)
		}
	}

	if len(pathsGenerated) == 0 || len(pathsGenerated) < len(spec.Paths) {
		ingress := g.newIngressResource(
			fmt.Sprintf("%s-ingress", opts.Service.Name),
			opts.Namespace,
			g.generatePath(&opts.Path, &opts.NGINXIngress),
			pathTypePrefix,
			g.generateAnnotations(&opts.Path, &opts.NGINXIngress, &opts.Ingress.CORS),
			&opts.Service,
			&opts.Ingress,
		)
		ingresses = append(ingresses, ingress)
	}

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
	ingressOpts *options.IngressOptions,
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
					Host: ingressOpts.Host,
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

func (g *Generator) shouldSplit(opts *options.Options, subOpts *options.SubOptions) bool {
	return !reflect.DeepEqual(opts.Ingress.CORS, subOpts.CORS)
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
