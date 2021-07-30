package nginx_ingress

import (
	"fmt"
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

func (g *Generator) Generate(options *options.Options, _ *openapi3.T) (string, error) {
	if err := options.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate options: %w", err)
	}

	ingress := v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ingressAPIVersion,
			Kind:       ingressKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-ingress", options.Service.Name),
			Namespace:   options.Namespace,
			Annotations: g.generateAnnotations(options),
		},
		Spec: v1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []v1.IngressRule{
				{
					Host: options.Ingress.Host,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									PathType: &pathTypePrefix,
									Path:     g.generatePath(options),
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: options.Service.Name,
											Port: v1.ServiceBackendPort{
												Number: options.Service.Port,
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

	b, err := yaml.Marshal(ingress)

	return string(b), err
}

func (g *Generator) generatePath(options *options.Options) string {
	if len(options.Path.TrimPrefix) > 0 &&
		strings.HasPrefix(options.Path.Base, options.Path.TrimPrefix) &&
		options.NGINXIngress.RewriteTarget == "" {
		pathSuffixRegex := "(/|$)(.*)"

		return options.Path.Base + pathSuffixRegex
	}

	return options.Path.Base
}

func (g *Generator) generateAnnotations(options *options.Options) map[string]string {
	rewriteTargetAnnotationKey := "nginx.ingress.kubernetes.io/rewrite-target"

	annotations := map[string]string{}

	if options.NGINXIngress.RewriteTarget != "" {
		annotations[rewriteTargetAnnotationKey] = options.NGINXIngress.RewriteTarget
	} else if len(options.Path.TrimPrefix) > 0 && strings.HasPrefix(options.Path.Base, options.Path.TrimPrefix) {
		annotations[rewriteTargetAnnotationKey] = "/$2"
	}

	return annotations
}
