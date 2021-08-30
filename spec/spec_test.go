package spec

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {

}

func TestParseFromReader(t *testing.T) {
	testCases := []struct {
		name   string
		spec   string
		result *openapi3.T
	}{
		{
			name: "swagger",
			spec: `swagger: "2.0"
info:
  title: Sample API
  description: API description in Markdown.
  version: 1.0.0
paths:
  /users:
    get: {}
`,
			result: &openapi3.T{
				OpenAPI: "3.0.3",
				Info: &openapi3.Info{
					Title:       "Sample API",
					Description: "API description in Markdown.",
					Version:     "1.0.0",
				},
				Paths: openapi3.Paths{
					"/users": &openapi3.PathItem{
						Get: &openapi3.Operation{},
					},
				},
			},
		},
		{
			name: "openapi",
			spec: `openapi: "3.0.3"
info:
  title: Sample API
  description: API description in Markdown.
  version: 1.0.0
paths:
  /users:
    get: {}
`,
			result: &openapi3.T{
				OpenAPI: "3.0.3",
				Info: &openapi3.Info{
					Title:       "Sample API",
					Description: "API description in Markdown.",
					Version:     "1.0.0",
				},
				Paths: openapi3.Paths{
					"/users": &openapi3.PathItem{
						Get: &openapi3.Operation{},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			actual, err := ParseFromReader(strings.NewReader(testCase.spec))
			r.NoError(err, "failed to parse spec from reader")
			r.Equal(testCase.result.OpenAPI, actual.OpenAPI)
			r.Equal(testCase.result.Info.Title, actual.Info.Title)
			r.Equal(testCase.result.Info.Description, actual.Info.Description)
			r.Equal(testCase.result.Info.Version, actual.Info.Version)
			r.NotNil(testCase.result.Paths.Find("/users"))
		})

	}
}
