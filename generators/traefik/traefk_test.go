package traefik

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/spec"
	"github.com/stretchr/testify/require"
)

func TestTraefik(t *testing.T) {
	var testCases = []struct {
		name    string
		options options.Options
		spec    string
		res     string
	}{}

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
