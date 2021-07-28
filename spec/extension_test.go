package spec

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/options"
)

func TestGetOptions(t *testing.T) {
	var testCases = []struct {
		name string
		spec *openapi3.T
		res  options.Options
	}{
		{
			name: "no extensions",
			spec: &openapi3.T{},
			res:  options.Options{},
		},
		{
			name: "global options set",
			spec: &openapi3.T{
				ExtensionProps: openapi3.ExtensionProps{
					Extensions: map[string]interface{}{
						kuskExtensionKey: json.RawMessage(`{"disabled":true}`),
					},
				},
			},
			res: options.Options{
				Disabled: true,
			},
		},
		{
			name: "path level options set",
			spec: &openapi3.T{
				Paths: openapi3.Paths{
					"/pet": &openapi3.PathItem{
						ExtensionProps: openapi3.ExtensionProps{
							Extensions: map[string]interface{}{
								kuskExtensionKey: json.RawMessage(`{"disabled":true}`),
							},
						},
					},
				},
			},
			res: options.Options{
				PathOperations: map[string]options.Options{
					"/pet": {
						Disabled: true,
					},
				},
			},
		},
		{
			name: "HTTP method level options set",
			spec: &openapi3.T{
				Paths: openapi3.Paths{
					"/pet": &openapi3.PathItem{
						Put: &openapi3.Operation{
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									kuskExtensionKey: json.RawMessage(`{"disabled":true}`),
								},
							},
						},
					},
				},
			},
			res: options.Options{
				PathOperations: map[string]options.Options{
					"/pet": {
						HTTPMethodOperations: map[string]options.Options{
							"PUT": {
								Disabled: true,
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			actual, err := GetOptions(testCase.spec)
			r.NoError(err, "failed to get options")
			r.Equal(testCase.res, *actual)
		})
	}

}
