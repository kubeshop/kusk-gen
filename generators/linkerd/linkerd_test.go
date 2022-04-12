package linkerd

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk-gen/options"
	"github.com/kubeshop/kusk-gen/spec"
)

type testCase struct {
	name    string
	options options.Options
	spec    string
	res     string
}

func TestLinkerd(t *testing.T) {
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

var trueValue = true
var falseValue = false

var testCases = []testCase{
	{
		name: "simple routes",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
		},
		spec: `openapi: 3.0.1
paths:
  /:
    get: {}

  /authors:
    post: {}
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
  - condition:
      method: POST
      pathRegex: /authors
    name: POST /authors
`,
	},
	{
		name: "simple routes with prefix",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
			Path: options.PathOptions{
				Base: "/prefix",
			},
		},
		spec: `openapi: 3.0.1
x-kusk:
  path:
    base: /prefix
paths:
  /:
    get: {}

  /authors:
    post: {}
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /prefix/
    name: GET /prefix/
  - condition:
      method: POST
      pathRegex: /prefix/authors
    name: POST /prefix/authors
`,
	},
	{
		name: "routes with variables",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
		},
		spec: `openapi: 3.0.1
paths:
  /:
    get: {}

  /books:
    post: {}

  /books/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64

  /books/{id}/edit:
    post:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
  - condition:
      method: GET
      pathRegex: /books/[^/]*
    name: GET /books/{id}
  - condition:
      method: POST
      pathRegex: /books
    name: POST /books
  - condition:
      method: POST
      pathRegex: /books/[^/]*/edit
    name: POST /books/{id}/edit
`,
	},
	{
		name: "path disabled",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
			PathSubOptions: map[string]options.SubOptions{
				"/books": {
					Disabled: &trueValue,
				},
			},
		},
		spec: `openapi: 3.0.1
paths:
  /:
    get: {}

  /books:
    x-kusk:
      disabled: true
    post: {}

  /books/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64

  /books/{id}/edit:
    post:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
  - condition:
      method: GET
      pathRegex: /books/[^/]*
    name: GET /books/{id}
  - condition:
      method: POST
      pathRegex: /books/[^/]*/edit
    name: POST /books/{id}/edit
`,
	},
	{
		name: "operation disabled",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
			OperationSubOptions: map[string]options.SubOptions{
				"GET/books": {
					Disabled: &trueValue,
				},
			},
		},
		spec: `openapi: 3.0.1
paths:
  /:
    get: {}

  /books:
    post: {}
    get:
      x-kusk:
        disabled: true

  /books/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64

  /books/{id}/edit:
    post:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
  - condition:
      method: GET
      pathRegex: /books/[^/]*
    name: GET /books/{id}
  - condition:
      method: POST
      pathRegex: /books
    name: POST /books
  - condition:
      method: POST
      pathRegex: /books/[^/]*/edit
    name: POST /books/{id}/edit
`,
	},
	{
		name: "simple routes with global timeout",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
			Timeouts: options.TimeoutOptions{
				RequestTimeout: 5,
			},
		},
		spec: `openapi: 3.0.1
paths:
  /:
    get: {}

  /authors:
    post: {}
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
    timeout: 5s
  - condition:
      method: POST
      pathRegex: /authors
    name: POST /authors
    timeout: 5s
`,
	},
	{
		name: "simple routes with path-level timeout",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
			Timeouts: options.TimeoutOptions{
				RequestTimeout: 5,
			},
			PathSubOptions: map[string]options.SubOptions{
				"/authors": {
					Timeouts: options.TimeoutOptions{
						RequestTimeout: 6,
					},
				},
			},
		},
		spec: `openapi: 3.0.1
x-kusk:
  timeouts:
    request_timeout: 5
paths:
  /:
    get: {}

  /authors:
    x-kusk:
      timeouts:
        request_timeout: 6
    post: {}
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
    timeout: 5s
  - condition:
      method: POST
      pathRegex: /authors
    name: POST /authors
    timeout: 6s
`,
	},
	{
		name: "simple routes with operation-level timeout",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: options.ClusterOptions{
				ClusterDomain: "cluster.local",
			},
			Timeouts: options.TimeoutOptions{
				RequestTimeout: 5,
			},
			OperationSubOptions: map[string]options.SubOptions{
				"POST/authors": {
					Timeouts: options.TimeoutOptions{
						RequestTimeout: 6,
					},
				},
			},
		},
		spec: `openapi: 3.0.1
x-kusk:
  timeouts:
    request_timeout: 5
paths:
  /:
    get: {}

  /authors:
    post:
      x-kusk:
        timeouts:
          request_timeout: 6
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
    timeout: 5s
  - condition:
      method: POST
      pathRegex: /authors
    name: POST /authors
    timeout: 6s
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
		name: "path disabled, operation enabled",
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
			},
			OperationSubOptions: map[string]options.SubOptions{
				"GET/": {
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
    get:
      x-kusk:
        disabled: false
    post: {}
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: petstore.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: GET
      pathRegex: /
    name: GET /
`,
	},
	{
		name: "path disabled not specified operation disabled specified",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "petstore",
			},
			OperationSubOptions: map[string]options.SubOptions{
				"GET/": {
					Disabled: &trueValue,
				},
				"POST/": {
					Disabled: &falseValue,
				},
				"PATCH/": {
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
    get:
      x-kusk:
        disabled: true
    post:
      x-kusk:
        disabled: false
    patch:
      x-kusk:
        disabled: false
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: petstore.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: PATCH
      pathRegex: /
    name: PATCH /
  - condition:
      method: POST
      pathRegex: /
    name: POST /
`,
	},
	{
		name: "path disabled not specified operation disabled specified operation enabled not specified",
		options: options.Options{
			Namespace: "default",
			Service: options.ServiceOptions{
				Namespace: "default",
				Name:      "petstore",
			},
			OperationSubOptions: map[string]options.SubOptions{
				"GET/": {
					Disabled: &trueValue,
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
    get:
      x-kusk:
        disabled: true
    post: {}
    patch: {}
`,
		res: `apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: petstore.default.svc.cluster.local
  namespace: default
spec:
  routes:
  - condition:
      method: PATCH
      pathRegex: /
    name: PATCH /
  - condition:
      method: POST
      pathRegex: /
    name: POST /
`,
	},
}
