package nginx_ingress

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	name    string
	options *Options
	res     string
}

func TestGenerate(t *testing.T) {
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
		options: &Options{
			ServiceName:      "webapp",
			ServiceNamespace: "default",
			Path:             "/",
			Port:             80,
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
		options: &Options{
			ServiceName:      "webapp",
			ServiceNamespace: "default",
			Path:             "/somepath",
			Port:             80,
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
		options: &Options{
			ServiceName:      "webapp",
			ServiceNamespace: "default",
			Path:             "/somepath",
			TrimPrefix:       "/somepath",
			Port:             80,
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
		options: &Options{
			ServiceName:      "webapp",
			ServiceNamespace: "default",
			Path:             "/somepath",
			TrimPrefix:       "/somepath",
			RewriteTarget:    "/someotherpath",
			Port:             80,
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
