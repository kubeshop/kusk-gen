package nginx_ingress

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/spec"
)

type testCase struct {
	name    string
	options options.Options
	spec    string
	res     string
}

func TestNGINXIngress(t *testing.T) {
	var testCases = []testCase{
		{
			name: "root base path and no trim prefix",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      80,
				},
				Path: options.PathOptions{
					Base: "/",
				},
			},
			res: `---
apiVersion: networking.k8s.io/v1
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
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      80,
				},
				Path: options.PathOptions{
					Base: "/somepath",
				},
			},
			res: `---
apiVersion: networking.k8s.io/v1
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
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      80,
				},
				Path: options.PathOptions{
					Base:       "/somepath",
					TrimPrefix: "/somepath",
				},
			},
			res: `---
apiVersion: networking.k8s.io/v1
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
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      80,
				},
				Path: options.PathOptions{
					Base:       "/somepath",
					TrimPrefix: "/somepath",
				},
				NGINXIngress: options.NGINXIngressOptions{
					RewriteTarget: "/someotherpath",
				},
			},
			res: `---
apiVersion: networking.k8s.io/v1
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
		{
			name: "CORS options set, nil credentials",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      80,
				},
				Path: options.PathOptions{
					Base:       "/somepath",
					TrimPrefix: "/somepath",
				},
				NGINXIngress: options.NGINXIngressOptions{
					RewriteTarget: "/someotherpath",
				},
				Ingress: options.IngressOptions{
					CORS: options.CORSOptions{
						Origins:       []string{"http://foo.example", "http://bar.example"},
						Methods:       []string{"POST", "GET", "OPTIONS"},
						Headers:       []string{"Content-Type"},
						ExposeHeaders: []string{"X-Custom-Header", "X-Other-Custom-Header"},
						Credentials:   nil,
						MaxAge:        120,
					},
				},
			},
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/cors-allow-headers: Content-Type
    nginx.ingress.kubernetes.io/cors-allow-methods: POST, GET, OPTIONS
    nginx.ingress.kubernetes.io/cors-allow-origin: http://foo.example
    nginx.ingress.kubernetes.io/cors-expose-headers: X-Custom-Header, X-Other-Custom-Header
    nginx.ingress.kubernetes.io/cors-max-age: "120"
    nginx.ingress.kubernetes.io/enable-cors: "true"
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

	var gen Generator

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)
			spec, err := spec.Parse([]byte(testCase.spec))
			r.NoError(err)
			profile, err := gen.Generate(&testCase.options, spec)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}

