package traefik

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	yaml "github.com/ghodss/yaml"
	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
	traefikDynamicConfig "github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefikCRD "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// This is default entrypoint, specified in static traefik configuration
	HTTPEntryPoint string = "web"
	APIVersion     string = "traefik.containo.us/v1alpha1"
	traefik        string = "traefik"
)

func init() {
	generators.Registry[traefik] = &Generator{}
}

type Generator struct{}

func (g *Generator) Cmd() string {
	return traefik
}

func (g *Generator) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet(traefik, pflag.ExitOnError)

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
		"host",
		"",
		"optional Host for Route match, e.g. example.org (default - match all hosts)",
	)
	return fs
}

func (g *Generator) ShortDescription() string {
	return "Generates Traefik resources"
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	if err := opts.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate opts: %w", err)
	}
	// TODO: do we really need this for Traefik?? Single Ingress looks ok.
	if opts.Path.Split {
		return generateSplitIngress(opts, spec)
	}
	return generateCombinedIngress(opts, spec)
}

func generateCombinedIngress(opts *options.Options, spec *openapi3.T) (string, error) {
	service := traefikCRD.Service{
		LoadBalancerSpec: traefikCRD.LoadBalancerSpec{
			Name:      opts.Service.Name,
			Namespace: opts.Service.Namespace,
			Port:      intstr.IntOrString{IntVal: opts.Service.Port},
		},
	}
	host := opts.Host
	base := opts.Path.Base
	// K8s name for created resources are based on service name
	name := opts.Service.Name
	// these middlewares will be used for all paths
	baseMiddlewares := generateBaseMiddlewares(name, opts)
	// these are all middlewares to create manifests
	// any new created middleware must go here
	allMiddlewares := append([]traefikCRD.Middleware{}, baseMiddlewares...)
	routes := []traefikCRD.Route{}
	// Iterate on all paths and build routes rules
	for path, pathItem := range spec.Paths {
		// x-kusk options per path
		pathSubOptions, ok := opts.PathSubOptions[path]
		if ok {
			if pathSubOptions.Disabled {
				continue
			}
		}

		if pathSubOptions.Host != "" && pathSubOptions.Host != host {
			host = pathSubOptions.Host
		}
		// take global CORS options
		corsOpts := opts.CORS

		// if non-zero path-level CORS options are different, override with them
		if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
			if !reflect.DeepEqual(options.CORSOptions{}, pathSubOpts.CORS) &&
				!reflect.DeepEqual(corsOpts, pathSubOpts.CORS) {
				corsOpts = pathSubOpts.CORS
			}
		}
		pathMiddlewares := []traefikCRD.Middleware{}
		// if final CORS options are not empty, include them
		if !reflect.DeepEqual(options.CORSOptions{}, corsOpts) {
			corsMiddleware := generateCORSMiddleware(name, path, opts, corsOpts)
			pathMiddlewares = append(pathMiddlewares, corsMiddleware)
			// Add to manifests generation
			allMiddlewares = append(allMiddlewares, pathMiddlewares...)
		}
		// x-kusk options per operation (http method)
		// For each method we create separate Match rule and route in case there are x-kusk extention overrides
		for method := range pathItem.Operations() {
			opSubOptions, ok := opts.OperationSubOptions[method+path]
			if ok {
				if opSubOptions.Disabled {
					continue
				}
				if opSubOptions.Host != "" && opSubOptions.Host != host {
					host = opSubOptions.Host
				}
			}

			matchRule := generateMatchRule(host, base, path, method)
			route := traefikCRD.Route{
				Match:    matchRule,
				Services: []traefikCRD.Service{service},
				Kind:     "Rule", Middlewares: generateMiddlewaresRefs(append(baseMiddlewares, pathMiddlewares...)),
			}
			routes = append(routes, route)
		}
	}
	ingressRouteSpec := traefikCRD.IngressRouteSpec{
		EntryPoints: []string{HTTPEntryPoint},
		Routes:      routes,
	}
	// This is finished Ingress object
	ingressRoute := traefikCRD.IngressRoute{
		Spec:       ingressRouteSpec,
		TypeMeta:   metav1.TypeMeta{Kind: "IngressRoute", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: opts.Service.Namespace},
	}
	return buildOutput(ingressRoute, allMiddlewares)
}

// TODO: write split ingress
func generateSplitIngress(opts *options.Options, spec *openapi3.T) (string, error) {
	return "", nil
}

// Separate CORS middleware
func generateCORSMiddleware(name string, path string, opts *options.Options, corsOpts options.CORSOptions) traefikCRD.Middleware {
	midlewareSpec := traefikCRD.MiddlewareSpec{Headers: &traefikDynamicConfig.Headers{
		AccessControlAllowHeaders:     corsOpts.Headers,
		AccessControlAllowMethods:     corsOpts.Methods,
		AccessControlAllowOriginList:  corsOpts.Origins,
		AccessControlMaxAge:           int64(corsOpts.MaxAge),
		AccessControlAllowCredentials: *corsOpts.Credentials,
	}}
	middleware := traefikCRD.Middleware{
		TypeMeta:   metav1.TypeMeta{Kind: "Middleware", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s-cors-headers", name, path), Namespace: opts.Namespace},
		Spec:       midlewareSpec,
	}
	return middleware
}

// Any non-overridable middlewares that apply to all paths
func generateBaseMiddlewares(name string, opts *options.Options) []traefikCRD.Middleware {
	var middlewares []traefikCRD.Middleware
	if opts.Path.TrimPrefix != "" {
		midlewareSpec := traefikCRD.MiddlewareSpec{StripPrefix: &traefikDynamicConfig.StripPrefix{Prefixes: []string{opts.Path.TrimPrefix}}}
		middleware := traefikCRD.Middleware{
			TypeMeta:   metav1.TypeMeta{Kind: "Middleware", APIVersion: APIVersion},
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-strip-base-prefix", name), Namespace: opts.Namespace},
			Spec:       midlewareSpec,
		}
		middlewares = append(middlewares, middleware)
	}
	return middlewares
}

func generateMiddlewaresRefs(middlewares []traefikCRD.Middleware) []traefikCRD.MiddlewareRef {
	var middlewaresRefs []traefikCRD.MiddlewareRef
	for _, m := range middlewares {
		middlewaresRefs = append(middlewaresRefs, traefikCRD.MiddlewareRef{Name: m.Name, Namespace: m.Namespace})
	}
	return middlewaresRefs
}

func generateMatchRule(host string, base string, path string, method string) string {
	const httpPathSeparator string = "/"
	// Avoids path joins (removes // in e.g. /path//subpath, or //subpath)
	fullPath := fmt.Sprintf(`%s/%s`, strings.TrimSuffix(base, httpPathSeparator), strings.TrimPrefix(path, httpPathSeparator))
	// Create rules to filter request on
	rules := []string{}
	// Host filter
	if host != "" {
		rules = append(rules, fmt.Sprintf("Host(`%s`)", host))
	}
	rules = append(rules, fmt.Sprintf("PathPrefix(`%s`)", fullPath))
	rules = append(rules, fmt.Sprintf("Method(`%s`)", method))
	// returns e.g. Host(`example.org`) && PathPrefix(`/petstore/api/v3/pet`) && Method(`POST`)
	return strings.Join(rules, " && ")
}

// Build suitable output to be piped into kubectl or a file
func buildOutput(ingressRoute traefikCRD.IngressRoute, middlewares []traefikCRD.Middleware) (string, error) {
	var builder strings.Builder
	// Middlewares first
	builder.WriteString("\n") // initial line feed
	for _, middleware := range middlewares {
		builder.WriteString("---\n") // indicate start of YAML resource
		b, err := yaml.Marshal(middleware)
		if err != nil {
			return "", fmt.Errorf("unable to marshal Middleware resource: %+v: %s", middleware, err.Error())
		}
		builder.WriteString(string(b))
	}
	// IngressRoute
	builder.WriteString("---\n") // indicate start of YAML resource
	b, err := yaml.Marshal(ingressRoute)
	if err != nil {
		return "", fmt.Errorf("unable to marshal IngressRoute resource: %+v: %s", ingressRoute, err.Error())
	}
	builder.WriteString(string(b))
	return builder.String(), nil
}
