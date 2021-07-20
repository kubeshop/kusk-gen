package nginx_ingress

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/generators"
)

type testCase struct {
	name    string
	options *generators.Options
	res     string
}

func TestNGINXIngress(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			profile, err := Generate(testCase.options, nil)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}

var testCases = []testCase{
	{
		name: "root base path and no trim prefix",
		options: &generators.Options{
			Namespace: "default",
			Service: &generators.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
				Port:      80,
			},
			Path: &generators.PathOptions{
				Base: "/",
			},
		},
		res: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  creationTimestamp: null
  name: webapp-ingress
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 80
        path: /
        pathType: Prefix
status:
  loadBalancer: {}
`,
	},
	{
		name: "non-root path and no trim prefix",
		options: &generators.Options{
			Namespace: "default",
			Service: &generators.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
				Port:      80,
			},
			Path: &generators.PathOptions{
				Base: "/somepath",
			},
		},
		res: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  creationTimestamp: null
  name: webapp-ingress
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 80
        path: /somepath
        pathType: Prefix
status:
  loadBalancer: {}
`,
	},
	{
		name: "non-root path and trim prefix",
		options: &generators.Options{
			Namespace: "default",
			Service: &generators.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
				Port:      80,
			},
			Path: &generators.PathOptions{
				Base:       "/somepath",
				TrimPrefix: "/somepath",
			},
		},
		res: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$2
  creationTimestamp: null
  name: webapp-ingress
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 80
        path: /somepath(/|$)(.*)
        pathType: Prefix
status:
  loadBalancer: {}
`,
	},
	{
		name: "non-root path and trim prefix and specified re-write target",
		options: &generators.Options{
			Namespace: "default",
			Service: &generators.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
				Port:      80,
			},
			Path: &generators.PathOptions{
				Base:       "/somepath",
				TrimPrefix: "/somepath",
			},
			NGINXIngress: &generators.NGINXIngressOptions{
				RewriteTarget: "/someotherpath",
			},
		},
		res: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /someotherpath
  creationTimestamp: null
  name: webapp-ingress
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 80
        path: /somepath
        pathType: Prefix
status:
  loadBalancer: {}
`,
	},
}
