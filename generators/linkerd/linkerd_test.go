package linkerd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/spec"
)

type testCase struct {
	name    string
	options *Options
	spec    string
	res     string
}

func TestLinkerd(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			spec, err := spec.Parse([]byte(testCase.spec))
			r.NoError(err, "failed to parse spec")

			profile, err := Generate(testCase.options, spec)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}

var testCases = []testCase{
	{
		name: "booksapp",
		options: &Options{
			Namespace:     "default",
			Name:          "webapp",
			ClusterDomain: "cluster.local",
		},
		spec: `
openapi: 3.0.1
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
}
