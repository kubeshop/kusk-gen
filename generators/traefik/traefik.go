package traefik

import (
	"fmt"
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
	HTTPEntryPoint  string = "web"
	APIVersion      string = "traefik.containo.us/v1alpha1"
	traefik         string = "traefik"
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

type ingressRouteData struct {
	Name             string
	Namespace        string
	Host             string
	PathBase         string
	PathTrimPrefix   string
	ServiceName      string
	ServiceNamespace string
	ServicePort      int32
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	if err := opts.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate opts: %w", err)
	}
	ingressData := ingressRouteData{
		Name:             opts.Service.Name,
		Namespace:        opts.Namespace,
		Host:             opts.Host,
		PathBase:         opts.Path.Base,
		PathTrimPrefix:   opts.Path.TrimPrefix,
		ServiceName:      opts.Service.Name,
		ServiceNamespace: opts.Service.Namespace,
		ServicePort:      opts.Service.Port,
	}

	service := traefikCRD.Service{
		LoadBalancerSpec: traefikCRD.LoadBalancerSpec{
			Name:      ingressData.ServiceName,
			Namespace: ingressData.ServiceNamespace,
			Port:      intstr.IntOrString{IntVal: ingressData.ServicePort},
		},
	}
	middlewares := generateMiddlewares(ingressData)
	route := traefikCRD.Route{
		Match:    generateMatchRule(ingressData),
		Services: []traefikCRD.Service{service},
		Kind:     "Rule", Middlewares: generateMiddlewaresRefs(middlewares),
	}
	ingressRouteSpec := traefikCRD.IngressRouteSpec{
		EntryPoints: []string{HTTPEntryPoint},
		Routes:      []traefikCRD.Route{route},
	}
	ingressRoute := traefikCRD.IngressRoute{
		Spec:       ingressRouteSpec,
		TypeMeta:   metav1.TypeMeta{Kind: "IngressRoute", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: ingressData.Name, Namespace: ingressData.Namespace},
	}
	return buildOutput(ingressRoute, middlewares)
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

func generateMiddlewares(ingressData ingressRouteData) []traefikCRD.Middleware {
	var middlewares []traefikCRD.Middleware
	if ingressData.PathTrimPrefix != "" {
		midlewareSpec := traefikCRD.MiddlewareSpec{StripPrefix: &traefikDynamicConfig.StripPrefix{Prefixes: []string{ingressData.PathTrimPrefix}}}
		middleware := traefikCRD.Middleware{
			TypeMeta:   metav1.TypeMeta{Kind: "Middleware", APIVersion: APIVersion},
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-strip-prefixes", ingressData.Name), Namespace: ingressData.Namespace},
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

func generateMatchRule(ingressData ingressRouteData) string {
	if ingressData.Host == "" {
		return fmt.Sprintf("PathPrefix(\"%s\")", ingressData.PathBase)
	} else {
		return fmt.Sprintf("Host(\"%s\") && PathPrefix(\"%s\")", ingressData.Host, ingressData.PathBase)
	}
}
