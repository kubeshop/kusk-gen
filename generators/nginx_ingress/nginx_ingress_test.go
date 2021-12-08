package nginx_ingress

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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
	trueValue := true
	falseValue := false

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
				CORS: options.CORSOptions{
					Origins:       []string{"http://foo.example", "http://bar.example"},
					Methods:       []string{"POST", "GET", "OPTIONS"},
					Headers:       []string{"Content-Type"},
					ExposeHeaders: []string{"X-Custom-Header", "X-Other-Custom-Header"},
					Credentials:   nil,
					MaxAge:        120,
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
		{
			name: "CORS options set differ at path level",
			options: options.Options{
				Namespace: "booksapp",
				Service: options.ServiceOptions{
					Namespace: "booksapp",
					Name:      "webapp",
					Port:      7000,
				},
				Path: options.PathOptions{
					Base:       "/bookstore",
					TrimPrefix: "/bookstore",
				},
				CORS: options.CORSOptions{
					Origins:       []string{"http://foo.example", "http://bar.example"},
					Methods:       []string{"POST", "GET", "OPTIONS"},
					Headers:       []string{"Content-Type"},
					ExposeHeaders: []string{"X-Custom-Header", "X-Other-Custom-Header"},
					Credentials:   nil,
					MaxAge:        86400,
				},
				PathSubOptions: map[string]options.SubOptions{
					"/books/{id}": {
						CORS: options.CORSOptions{
							Methods:       []string{"POST", "GET", "OPTIONS"},
							Headers:       []string{"Content-Type"},
							ExposeHeaders: []string{"X-Custom-Header"},
							Credentials:   nil,
							MaxAge:        86400,
						},
					},
				},
			},
			spec: `openapi: 3.0.1
x-kusk:
  namespace: booksapp
  path:
    base: /bookstore
    trim_prefix: /bookstore
  cors:
    origins:
      - http://foo.example
      - http://bar.example
    methods:
      - POST
      - GET
      - OPTIONS
    headers:
      - Content-Type
    credentials: true
    expose_headers:
      - X-Custom-Header
      - X-Other-Custom-Header
    max_age: 86400
  service:
    name: webapp
    namespace: booksapp
    port: 7000
paths:
  /:
    get: {}

  /books/{id}:
    x-kusk:
      cors:
        methods:
          - POST
          - GET
          - OPTIONS
        headers:
          - Content-Type
        credentials: true
        expose_headers:
          - X-Custom-Header
        max_age: 86400
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
`,
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/cors-allow-headers: Content-Type
    nginx.ingress.kubernetes.io/cors-allow-methods: POST, GET, OPTIONS
    nginx.ingress.kubernetes.io/cors-expose-headers: X-Custom-Header
    nginx.ingress.kubernetes.io/cors-max-age: "86400"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/rewrite-target: /books/$1
    nginx.ingress.kubernetes.io/use-regex: "true"
  creationTimestamp: null
  name: webapp-books-id
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 7000
        path: /bookstore/books/([A-z0-9]+)
        pathType: Exact
status:
  loadBalancer: {}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/cors-allow-headers: Content-Type
    nginx.ingress.kubernetes.io/cors-allow-methods: POST, GET, OPTIONS
    nginx.ingress.kubernetes.io/cors-allow-origin: http://foo.example
    nginx.ingress.kubernetes.io/cors-expose-headers: X-Custom-Header, X-Other-Custom-Header
    nginx.ingress.kubernetes.io/cors-max-age: "86400"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/use-regex: "true"
  creationTimestamp: null
  name: webapp-root
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 7000
        path: /bookstore$
        pathType: Exact
status:
  loadBalancer: {}
`,
		},
		{
			name: "request timeout options",
			options: options.Options{
				Namespace: "booksapp",
				Service: options.ServiceOptions{
					Namespace: "booksapp",
					Name:      "webapp",
					Port:      7000,
				},
				Path: options.PathOptions{
					Base:       "/bookstore",
					TrimPrefix: "/bookstore",
				},
				Timeouts: options.TimeoutOptions{
					RequestTimeout: 10,
					IdleTimeout:    0,
				},
			},
			spec: `openapi: 3.0.1
x-kusk:
  namespace: booksapp
  timeouts:
    request_timeout: 10
  path:
    base: /bookstore
    trim_prefix: /bookstore
  service:
    name: webapp
    namespace: booksapp
    port: 7000
paths:
  /:
    get: {}

  /books/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
`,
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "5"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "5"
    nginx.ingress.kubernetes.io/rewrite-target: /$2
  creationTimestamp: null
  name: webapp-ingress
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 7000
        path: /bookstore(/|$)(.*)
        pathType: Prefix
status:
  loadBalancer: {}
`,
		},
		{
			name: "rate limit options",
			options: options.Options{
				Namespace: "booksapp",
				Service: options.ServiceOptions{
					Namespace: "booksapp",
					Name:      "webapp",
					Port:      7000,
				},
				Path: options.PathOptions{
					Base:       "/bookstore",
					TrimPrefix: "/bookstore",
				},
				RateLimits: options.RateLimitOptions{
					RPS:   100,
					Burst: 400,
				},
			},
			spec: `openapi: 3.0.1
x-kusk:
  namespace: booksapp
  rate_limits:
    rps: 100
    burst: 400
  path:
    base: /bookstore
    trim_prefix: /bookstore
  service:
    name: webapp
    namespace: booksapp
    port: 7000
paths:
  /:
    get: {}

  /books/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
`,
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/limit-burst-multiplier: "4"
    nginx.ingress.kubernetes.io/limit-rps: "100"
    nginx.ingress.kubernetes.io/rewrite-target: /$2
  creationTimestamp: null
  name: webapp-ingress
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: webapp
            port:
              number: 7000
        path: /bookstore(/|$)(.*)
        pathType: Prefix
status:
  loadBalancer: {}
`,
		},
		{
			name: "globally disabled",
			options: options.Options{
				Disabled:  true,
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "petstore",
				},
			},
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  disabled: true
paths:
  /:
    get: {}
    post: {}
`,
			res: ``,
		},
		{
			name: "path disabled, another path enabled",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "petstore",
				},
				PathSubOptions: map[string]options.SubOptions{
					"/": {
						Disabled: &trueValue,
					},
					"/path": {
						Disabled: &falseValue,
					},
				},
			},
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
paths:
  /:
    x-kusk:
      disabled: true
    get: {}
  /path:
    x-kusk:
      disabled: true
    get: {}
`,
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /path
  creationTimestamp: null
  name: petstore-path
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: petstore
            port:
              number: 80
        path: /path
        pathType: Exact
status:
  loadBalancer: {}
`,
		},
		{
			name: "globally disabled, single path enabled",
			options: options.Options{
				Disabled:  true,
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "petstore",
				},
				PathSubOptions: map[string]options.SubOptions{
					"/path": {
						Disabled: &falseValue,
					},
				},
			},
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  disabled: true
paths:
  /:
    get: {}
  /path:
    x-kusk:
      disabled: false
    get: {}
`,
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /path
  creationTimestamp: null
  name: petstore-path
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: petstore
            port:
              number: 80
        path: /path
        pathType: Exact
status:
  loadBalancer: {}
`,
		},
		{
			name: "REGRESSION - rewrite target correct when base path defined and trim prefix is not defined",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "petstore",
				},
				Path: options.PathOptions{
					Base: "/api",
				},
				PathSubOptions: map[string]options.SubOptions{
					"/": {
						Disabled: &trueValue,
					},
				},
			},
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  service:
    name: petstore
    namespace: default
  path:
    base: /api
paths:
  /:
    x-kusk:
      disabled: true
    get: {}
  /path:
    get: {}
`,
			res: `---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /api/path
  creationTimestamp: null
  name: petstore-path
  namespace: default
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: petstore
            port:
              number: 80
        path: /api/path
        pathType: Exact
status:
  loadBalancer: {}
`,
		},
	}

	var gen Generator

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			spec, err := spec.NewParser(openapi3.NewLoader()).ParseFromReader(strings.NewReader(testCase.spec))
			r.NoError(err)
			profile, err := gen.Generate(&testCase.options, spec)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}
