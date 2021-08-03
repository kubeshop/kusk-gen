package nginx_ingress

import (
	"fmt"
	"reflect"
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

	for path, subOpts := range opts.PathSubOptions {
		if g.shouldSplit(opts, &subOpts) {
			fmt.Printf("should split on path: %s", path)
		}
	}

	ingresses = append(ingresses, v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ingressAPIVersion,
			Kind:       ingressKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-ingress", opts.Service.Name),
			Namespace:   opts.Namespace,
			Annotations: g.generateAnnotations(&opts.Path, &opts.NGINXIngress, &opts.Ingress.CORS),
		},
		Spec: v1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []v1.IngressRule{
				{
					Host: opts.Ingress.Host,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									PathType: &pathTypePrefix,
									Path:     g.generatePath(&opts.Path, &opts.NGINXIngress),
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: opts.Service.Name,
											Port: v1.ServiceBackendPort{
												Number: opts.Service.Port,
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
	})

	b, err := yaml.Marshal(ingresses)

	return string(b), err
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
