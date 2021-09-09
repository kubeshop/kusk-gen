package traefik

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/spec"
)

func TestTraefik(t *testing.T) {
	var testCases = []struct {
		name    string
		options options.Options
		spec    string
		res     string
	}{
		{
			name: "root base path and no trim prefix",
			options: options.Options{
				Namespace: "default",
				Path: options.PathOptions{
					Base: "/",
				},
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "petstore",
					Port:      7000,
				},
			},
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: notdefault
  service:
    name: petstore
paths:
  "/pet":
    put:
      operationId: updatePet
      responses:
        '200':
          description: Successful operation
`,
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: default
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/")
    services:
    - name: petstore
      namespace: default
      port: 7000
`,
		},
		{
			name: "non-root path and no trim prefix",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      7000,
				},
				Path: options.PathOptions{
					Base: "/somepath",
				},
			},
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: webapp
  namespace: default
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/somepath")
    services:
    - name: webapp
      namespace: default
      port: 7000
`,
		},
		{

			name: "non-root path and trim prefix",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      7000,
				},
				Path: options.PathOptions{
					Base:       "/somepath",
					TrimPrefix: "/somepath",
				},
			},
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: webapp-strip-prefixes
  namespace: default
spec:
  stripPrefix:
    prefixes:
    - /somepath
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: webapp
  namespace: default
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/somepath")
    middlewares:
    - name: webapp-strip-prefixes
      namespace: default
    services:
    - name: webapp
      namespace: default
      port: 7000
`,
		},
		{
			name: "non-root path and host match",
			options: options.Options{
				Namespace: "default",
				Service: options.ServiceOptions{
					Namespace: "default",
					Name:      "webapp",
					Port:      7000,
				},
				Path: options.PathOptions{
					Base: "/somepath",
				},
				Host: "example.com",
			},
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: webapp
  namespace: default
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: Host("example.com") && PathPrefix("/somepath")
    services:
    - name: webapp
      namespace: default
      port: 7000
`,
		},
	}

	var gen Generator

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			spec, err := spec.NewParser(openapi3.NewLoader()).ParseFromReader(strings.NewReader(testCase.spec))
			r.NoError(err, "failed to parse spec")

			profile, err := gen.Generate(&testCase.options, spec)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}
