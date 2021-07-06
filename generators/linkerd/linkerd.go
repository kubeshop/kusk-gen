package linkerd

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/linkerd/linkerd2/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Options struct {
	Namespace     string
	Name          string
	ClusterDomain string
}

func Generate(options *Options, spec *openapi3.T) (string, error) {
	profile := &v1alpha2.ServiceProfile{
		TypeMeta: metav1.TypeMeta{
			APIVersion: k8s.ServiceProfileAPIVersion,
			Kind:       k8s.ServiceProfileKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf(
				"%s.%s.svc.%s",
				options.Name,
				options.Namespace,
				options.ClusterDomain,
				),
			Namespace: options.Namespace,
		},
		Spec: generateProfileSpec(spec),
	}

	b, err := yaml.Marshal(profile)

	return string(b), err
}

func generateProfileSpec(spec *openapi3.T) v1alpha2.ServiceProfileSpec {
	routes := make([]*v1alpha2.RouteSpec, 0)

	for path, pathItem := range spec.Paths {
		for method := range pathItem.Operations() {
			routes = append(routes, generateRouteSpec(method, path))
		}
	}

	return v1alpha2.ServiceProfileSpec{Routes: routes}
}

func generateRouteSpec(method, path string) *v1alpha2.RouteSpec {
	return &v1alpha2.RouteSpec{
		Name:            fmt.Sprintf("%s %s", method, path),
		Condition:       &v1alpha2.RequestMatch{
			PathRegex: path,
			Method:    method,
		},
	}
}