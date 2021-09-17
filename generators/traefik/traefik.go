package traefik

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	"github.com/ghodss/yaml"
	traefikDynamicConfig "github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefikCRD "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
)

const (
	// This is default entrypoint, specified in static traefik configuration
	HTTPEntryPoint string = "web"
	APIVersion     string = "traefik.containo.us/v1alpha1"
	traefik        string = "traefik"
)

var (
	rePathSymbols = regexp.MustCompile(`[/{}]`)
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

	fs.Uint32(
		"rate_limits.rps",
		0,
		"request per second rate limit",
	)

	fs.Uint32(
		"rate_limits.burst",
		0,
		"request per second burst",
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

	fs.String(
		"host",
		"",
		"the Host header value to listen on",
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
	host := opts.Host
	base := opts.Path.Base
	// K8s serviceName for created resources are based on service serviceName
	serviceName := opts.Service.Name
	namespace := opts.Namespace

	// these are all middlewares to create manifests for
	allMiddlewares := []traefikCRD.Middleware{}

	// Map to populate middlewares for each route
	// We use it to filter out duplicates (e.g. multiple CORS specifications per path, method)
	rootMiddlewares := map[string]traefikCRD.Middleware{}

	if opts.Path.TrimPrefix != "" {
		stripPrefixMiddleware := generateStripPrefixMiddleware(generateResourceName([]string{serviceName, "strip-prefix"}), namespace, opts.Path.TrimPrefix)
		rootMiddlewares["stripprefix"] = stripPrefixMiddleware
		allMiddlewares = append(allMiddlewares, stripPrefixMiddleware)
	}

	// Top level CORS middleware, could be overriden per path/method
	if !reflect.DeepEqual(options.CORSOptions{}, opts.CORS) {
		corsMiddleware := generateCORSMiddleware(generateResourceName([]string{serviceName, "cors"}), namespace, opts.CORS)
		rootMiddlewares["cors"] = corsMiddleware
		allMiddlewares = append(allMiddlewares, corsMiddleware)
	}

	// Top level RateLimit middleware
	if !reflect.DeepEqual(options.RateLimitOptions{}, opts.RateLimits) {
		rateLimitMiddleware := generateRateLimitMiddleware(generateResourceName([]string{serviceName, "ratelimit"}), namespace, opts.RateLimits)
		rootMiddlewares["ratelimit"] = rateLimitMiddleware
		allMiddlewares = append(allMiddlewares, rateLimitMiddleware)
	}
	// Default top level service servers transport (defines communication with service backend, e.g. timeouts, tls)
	serviceServersTransport := generateServerTransport(serviceName, namespace, opts.Timeouts)
	allServersTransports := []traefikCRD.ServersTransport{serviceServersTransport}

	// Routes to include into ingress
	routes := []traefikCRD.Route{}

	// Main routine
	// Iterate on all paths and build routes rules with related middlewares and any overrides
	for path, pathItem := range spec.Paths {
		// x-kusk options per operation (http method)
		// For each method we create separate Match rule and route and then add to routes list
		for method := range pathItem.Operations() {
			if opts.IsOperationDisabled(path, method) {
				continue
			}

			// x-kusk options per path
			// ServersTransport for this path
			pathServiceServersTransport := serviceServersTransport
			// Create copy of root middlewares map to further override per path
			pathMiddlewares := copyMiddlewareMap(rootMiddlewares)
			// x-kusk options per path
			pathSubOpts, ok := opts.PathSubOptions[path]
			if ok {
				if pathSubOpts.Host != "" && pathSubOpts.Host != host {
					host = pathSubOpts.Host
				}
				// if non-zero path-level CORS options are different, override with them
				if !reflect.DeepEqual(options.CORSOptions{}, pathSubOpts.CORS) {
					corsMiddleware := generateCORSMiddleware(generateResourceName([]string{serviceName, path, "cors"}), namespace, pathSubOpts.CORS)
					pathMiddlewares["cors"] = corsMiddleware
					allMiddlewares = append(allMiddlewares, corsMiddleware)
				}

				if !reflect.DeepEqual(options.RateLimitOptions{}, pathSubOpts.RateLimits) {
					rateLimitMiddleware := generateRateLimitMiddleware(generateResourceName([]string{serviceName, path, "ratelimit"}), namespace, pathSubOpts.RateLimits)
					pathMiddlewares["ratelimit"] = rateLimitMiddleware
					allMiddlewares = append(allMiddlewares, rateLimitMiddleware)
				}

				if !reflect.DeepEqual(options.TimeoutOptions{}, pathSubOpts.Timeouts) {
					pathServiceServersTransport = generateServerTransport(generateResourceName([]string{serviceName, path}), opts.Namespace, pathSubOpts.Timeouts)
					allServersTransports = append(allServersTransports, pathServiceServersTransport)
				}
			}

			// Create copy of path middlewares map to further override per method
			opMiddlewares := copyMiddlewareMap(pathMiddlewares)

			// We override any suboptions
			opSubOpts, ok := opts.OperationSubOptions[method+path]
			opServiceServersTransport := pathServiceServersTransport
			if ok {
				if opts.IsOperationDisabled(path, method) {
					continue
				}
				if opSubOpts.Host != "" && opSubOpts.Host != host {
					host = opSubOpts.Host
				}
				// if non-zero operation level CORS options are different, override with them
				if !reflect.DeepEqual(options.CORSOptions{}, opSubOpts.CORS) {
					corsMiddleware := generateCORSMiddleware(generateResourceName([]string{serviceName, path, method, "cors"}), namespace, opSubOpts.CORS)
					opMiddlewares["cors"] = corsMiddleware
					allMiddlewares = append(allMiddlewares, corsMiddleware)
				}

				if !reflect.DeepEqual(options.RateLimitOptions{}, opSubOpts.RateLimits) {
					rateLimitMiddleware := generateRateLimitMiddleware(generateResourceName([]string{serviceName, path, "ratelimit"}), namespace, opSubOpts.RateLimits)
					opMiddlewares["ratelimit"] = rateLimitMiddleware
					allMiddlewares = append(allMiddlewares, rateLimitMiddleware)
				}

				if !reflect.DeepEqual(options.TimeoutOptions{}, opSubOpts.Timeouts) {
					opServiceServersTransport = generateServerTransport(generateResourceName([]string{serviceName, path, method}), namespace, opSubOpts.Timeouts)
					allServersTransports = append(allServersTransports, opServiceServersTransport)
				}
			}

			matchRule := generateMatchRule(host, base, path, method)
			service := traefikCRD.Service{
				LoadBalancerSpec: traefikCRD.LoadBalancerSpec{
					Name:             serviceName,
					Namespace:        namespace,
					Port:             intstr.IntOrString{IntVal: opts.Service.Port},
					ServersTransport: opServiceServersTransport.ObjectMeta.Name,
				},
			}
			route := traefikCRD.Route{
				Match:       matchRule,
				Services:    []traefikCRD.Service{service},
				Kind:        "Rule",
				Middlewares: generateMiddlewaresRefs(middlewareMapToList(opMiddlewares)),
			}
			routes = append(routes, route)
		}
	}

	if len(routes) == 0 {
		return "", nil
	}

	// Finally generate Ingress spec and object itself
	ingressRouteSpec := traefikCRD.IngressRouteSpec{
		EntryPoints: []string{HTTPEntryPoint},
		Routes:      routes,
	}

	// Sort the list for tests to be stable
	sort.SliceStable(routes, func(i, j int) bool {
		return routes[i].Match < routes[j].Match
	})
	ingressRoute := traefikCRD.IngressRoute{
		Spec:       ingressRouteSpec,
		TypeMeta:   metav1.TypeMeta{Kind: "IngressRoute", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: namespace},
	}
	return buildOutput(ingressRoute, allMiddlewares, allServersTransports)
}

func generateCORSMiddleware(name string, namespace string, corsOpts options.CORSOptions) traefikCRD.Middleware {
	midlewareSpec := traefikCRD.MiddlewareSpec{Headers: &traefikDynamicConfig.Headers{
		AccessControlAllowHeaders:     corsOpts.Headers,
		AccessControlAllowMethods:     corsOpts.Methods,
		AccessControlAllowOriginList:  corsOpts.Origins,
		AccessControlMaxAge:           int64(corsOpts.MaxAge),
		AccessControlAllowCredentials: *corsOpts.Credentials,
	}}
	middleware := traefikCRD.Middleware{
		TypeMeta:   metav1.TypeMeta{Kind: "Middleware", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       midlewareSpec,
	}
	return middleware
}

func generateRateLimitMiddleware(name string, namespace string, rateLimitOpts options.RateLimitOptions) traefikCRD.Middleware {
	burst := int64(rateLimitOpts.Burst)
	midlewareSpec := traefikCRD.MiddlewareSpec{
		RateLimit: &traefikCRD.RateLimit{
			Average: int64(rateLimitOpts.RPS),
			Burst:   &burst,
		},
	}
	middleware := traefikCRD.Middleware{
		TypeMeta:   metav1.TypeMeta{Kind: "Middleware", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       midlewareSpec,
	}
	return middleware
}

func generateStripPrefixMiddleware(name string, namespace string, prefix string) traefikCRD.Middleware {
	midlewareSpec := traefikCRD.MiddlewareSpec{StripPrefix: &traefikDynamicConfig.StripPrefix{Prefixes: []string{prefix}}}
	middleware := traefikCRD.Middleware{
		TypeMeta:   metav1.TypeMeta{Kind: "Middleware", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       midlewareSpec,
	}
	return middleware
}

func generateServerTransport(name string, namespace string, timeouts options.TimeoutOptions) traefikCRD.ServersTransport {
	return traefikCRD.ServersTransport{
		TypeMeta:   metav1.TypeMeta{Kind: "ServersTransport", APIVersion: APIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: traefikCRD.ServersTransportSpec{
			ForwardingTimeouts: &traefikCRD.ForwardingTimeouts{
				IdleConnTimeout:       &intstr.IntOrString{IntVal: int32(timeouts.IdleTimeout)},
				ResponseHeaderTimeout: &intstr.IntOrString{IntVal: int32(timeouts.RequestTimeout)},
				DialTimeout:           &intstr.IntOrString{IntVal: int32(timeouts.RequestTimeout)},
			},
		},
	}
}

func generateMiddlewaresRefs(middlewares []traefikCRD.Middleware) []traefikCRD.MiddlewareRef {
	middlewaresRefs := []traefikCRD.MiddlewareRef{}
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
		rules = append(rules, fmt.Sprintf("Host(\"%s\")", host))
	}
	rules = append(rules, fmt.Sprintf("PathPrefix(\"%s\")", fullPath))
	rules = append(rules, fmt.Sprintf("Method(\"%s\")", method))
	// returns e.g. Host(`example.org`) && PathPrefix(`/petstore/api/v3/pet`) && Method(`POST`)
	return strings.Join(rules, " && ")
}

func generateResourceName(s []string) string {
	//sanitize
	for i := 0; i < len(s); i++ {
		s[i] = rePathSymbols.ReplaceAllString(s[i], "")
	}
	return strings.ToLower(strings.Join(s, "-"))
}

// Build suitable output to be piped into kubectl or a file
func buildOutput(ingressRoute traefikCRD.IngressRoute, middlewares []traefikCRD.Middleware, serversTransports []traefikCRD.ServersTransport) (string, error) {
	var builder strings.Builder
	// Middlewares first
	builder.WriteString("\n") // initial line feed

	// Sort the list for tests to be stable
	sort.SliceStable(middlewares, func(i, j int) bool {
		return middlewares[i].ObjectMeta.Name < middlewares[j].ObjectMeta.Name
	})
	for _, middleware := range middlewares {
		builder.WriteString("---\n") // indicate start of YAML resource
		b, err := yaml.Marshal(middleware)
		if err != nil {
			return "", fmt.Errorf("unable to marshal Middleware resource: %+v: %s", middleware, err.Error())
		}
		builder.WriteString(string(b))
	}
	for _, serversTransport := range serversTransports {
		builder.WriteString("---\n") // indicate start of YAML resource
		b, err := yaml.Marshal(serversTransport)
		if err != nil {
			return "", fmt.Errorf("unable to marshal ServersTransport resource: %+v: %s", serversTransport, err.Error())
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

func middlewareMapToList(m map[string]traefikCRD.Middleware) []traefikCRD.Middleware {
	l := []traefikCRD.Middleware{}
	for _, v := range m {
		l = append(l, v)
	}
	// Sort the list for tests since items in the map are unsorted
	sort.SliceStable(l, func(i, j int) bool {
		return l[i].ObjectMeta.Name < l[j].ObjectMeta.Name
	})
	return l
}

func copyMiddlewareMap(m map[string]traefikCRD.Middleware) map[string]traefikCRD.Middleware {
	resMap := map[string]traefikCRD.Middleware{}
	for k, v := range m {
		resMap[k] = v
	}
	return resMap
}
