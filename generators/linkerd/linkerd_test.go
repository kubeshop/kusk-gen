package linkerd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/spec"
)

type testCase struct {
	name    string
	options generators.Options
	spec    string
	res     string
}

func TestLinkerd(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			spec, err := spec.Parse([]byte(testCase.spec))
			r.NoError(err, "failed to parse spec")

			profile, err := Generate(&testCase.options, spec)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}

var testCases = []testCase{
	{
		name: "simple routes",
		options: generators.Options{
			Namespace: "default",
			Service: generators.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: generators.ClusterOptions{
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
		name: "routes with variables",
		options: generators.Options{
			Namespace: "default",
			Service: generators.ServiceOptions{
				Namespace: "default",
				Name:      "webapp",
			},
			Cluster: generators.ClusterOptions{
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
}
