package linkerd

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/linkerd/linkerd2/pkg/k8s"
	"github.com/linkerd/linkerd2/pkg/profiles"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshop/kusk/generators"
)

func Generate(options generators.Options, spec *openapi3.T) (string, error) {
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
		Spec: generateServiceProfileSpec(spec),
	}

	b, err := yaml.Marshal(profile)

	return string(b), err
}

func generateServiceProfileSpec(spec *openapi3.T) v1alpha2.ServiceProfileSpec {
	routes := make([]*v1alpha2.RouteSpec, 0)

	for path, pathItem := range spec.Paths {
		for method := range pathItem.Operations() {
			routes = append(routes, generateRouteSpec(method, path))
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Name < routes[j].Name
	})

	return v1alpha2.ServiceProfileSpec{Routes: routes}
}

func generateRouteSpec(method, path string) *v1alpha2.RouteSpec {
	return &v1alpha2.RouteSpec{
		Name: fmt.Sprintf("%s %s", method, path),
		Condition: &v1alpha2.RequestMatch{
			PathRegex: profiles.PathToRegex(path),
			Method:    method,
		},
	}
}
