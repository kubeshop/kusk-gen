package nginxIngress

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressAPIVersion = "networking.k8s.io/v1"
	ingressKind       = "Ingress"
)

var (
	ingressClassName = "nginx"
	pathTypePrefix   = v1.PathTypePrefix
)

type Options struct {
	ServiceName      string
	ServiceNamespace string
	Path             string
	Port             int32
}

func Generate(options *Options, _ *openapi3.T) (string, error) {

	ingress := v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ingressAPIVersion,
			Kind:       ingressKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress", options.ServiceName),
			Namespace: options.ServiceNamespace,
			Annotations: map[string]string{
				//"kubernetes.io/ingress.class":                       ingressClassName,
				"nginxIngress.ingress.kubernetes.io/rewrite-target": "/\\$1",
			},
		},
		Spec: v1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []v1.IngressRule{
				{
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									PathType: &pathTypePrefix,
									Path:     options.Path,
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: options.ServiceName,
											Port: v1.ServiceBackendPort{
												Number: options.Port,
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
