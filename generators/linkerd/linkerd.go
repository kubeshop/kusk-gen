package linkerd

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/linkerd/linkerd2/pkg/k8s"
	"github.com/linkerd/linkerd2/pkg/profiles"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
)

func init() {
	generators.Registry["linkerd"] = &Generator{}
}

type Generator struct{}

func (g *Generator) Cmd() string {
	return "linkerd"
}

func (g *Generator) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("linkerd", pflag.ExitOnError)

	fs.String(
		"cluster.cluster_domain",
		"cluster.local",
		"kubernetes cluster domain",
	)

	fs.String(
		"path.base",
		"/",
		"a base prefix for Service endpoints",
	)

	fs.Uint32(
		"timeouts.request_timeout",
		0,
		"total request timeout (seconds)",
	)

	return fs
}

func (g *Generator) ShortDescription() string {
	return "Generates Linkerd Service Profiles for your service"
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Generate(options *options.Options, spec *openapi3.T) (string, error) {
	if err := options.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate options: %w", err)
	}

	spSpec := g.generateServiceProfileSpec(options, spec)
	if len(spSpec.Routes) == 0 {
		return "", nil
	}

	profile := &v1alpha2.ServiceProfile{
		TypeMeta: metav1.TypeMeta{
			APIVersion: k8s.ServiceProfileAPIVersion,
			Kind:       k8s.ServiceProfileKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf(
				"%s.%s.svc.%s",
				options.Service.Name,
				options.Service.Namespace,
				options.Cluster.ClusterDomain,
			),
			Namespace: options.Namespace,
		},
		Spec: spSpec,
	}

	b, err := yaml.Marshal(profile)

	return string(b), err
}

func (g *Generator) generateServiceProfileSpec(options *options.Options, spec *openapi3.T) v1alpha2.ServiceProfileSpec {
	routes := make([]*v1alpha2.RouteSpec, 0)

	for path, pathItem := range spec.Paths {
		for method, _ := range pathItem.Operations() {
			if options.IsOperationDisabled(path, method) {
				continue
			}

			routes = append(routes, generateRouteSpec(method, path, options))
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Name < routes[j].Name
	})

	return v1alpha2.ServiceProfileSpec{Routes: routes}
}

func generateRouteSpec(method, path string, opts *options.Options) *v1alpha2.RouteSpec {
	path = strings.TrimSuffix(opts.Path.Base, "/") + "/" + strings.TrimPrefix(path, "/")

	res := &v1alpha2.RouteSpec{
		Name: fmt.Sprintf("%s %s", method, path),
		Condition: &v1alpha2.RequestMatch{
			PathRegex: profiles.PathToRegex(path),
			Method:    method,
		},
	}

	// global timeouts are defined, use them
	if !reflect.DeepEqual(options.TimeoutOptions{}, opts.Timeouts) {
		res.Timeout = formatTimeout(opts.Timeouts.RequestTimeout)
	}

	// non-zero path-level timeouts are defined, use them
	if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
		if !reflect.DeepEqual(options.TimeoutOptions{}, pathSubOpts.Timeouts) {
			res.Timeout = formatTimeout(pathSubOpts.Timeouts.RequestTimeout)
		}
	}

	// non-zero operation-level timeouts are defined, use them
	if operationSubOpts, ok := opts.OperationSubOptions[method+path]; ok {
		if !reflect.DeepEqual(options.TimeoutOptions{}, operationSubOpts.Timeouts) {
			res.Timeout = formatTimeout(operationSubOpts.Timeouts.RequestTimeout)
		}
	}

	return res
}

func formatTimeout(timeout uint32) string {
	return fmt.Sprintf("%ds", timeout)
}
