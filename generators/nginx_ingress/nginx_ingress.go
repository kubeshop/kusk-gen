package nginx_ingress

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshop/kusk/generators"
)

const (
	ingressAPIVersion = "networking.k8s.io/v1"
	ingressKind       = "Ingress"
)

var (
	ingressClassName = "nginx"
	pathTypePrefix   = v1.PathTypePrefix
)

func Generate(options *generators.Options, _ *openapi3.T) (string, error) {
	ingress := v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ingressAPIVersion,
			Kind:       ingressKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-ingress", options.Service.Name),
			Namespace:   options.Namespace,
			Annotations: generateAnnotations(options),
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
									Path:     generatePath(options),
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

func generatePath(options *generators.Options) string {
	if len(options.Path.TrimPrefix) > 0 &&
		strings.HasPrefix(options.Path.Base, options.Path.TrimPrefix) &&
		options.NGINXIngress.RewriteTarget == "" {
		pathSuffixRegex := "(/|$)(.*)"

		return options.Path.Base + pathSuffixRegex
	}

	return options.Path.Base
}

func generateAnnotations(options *generators.Options) map[string]string {
	rewriteTargetAnnotationKey := "nginx.ingress.kubernetes.io/rewrite-target"

	annotations := map[string]string{}

	if options.NGINXIngress.RewriteTarget != "" {
		annotations[rewriteTargetAnnotationKey] = options.NGINXIngress.RewriteTarget
	} else if len(options.Path.TrimPrefix) > 0 && strings.HasPrefix(options.Path.Base, options.Path.TrimPrefix) {
		annotations[rewriteTargetAnnotationKey] = "/$2"
	}

	return annotations
}
