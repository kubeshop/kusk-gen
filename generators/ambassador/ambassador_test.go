package ambassador

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func parseSpec(t *testing.T, spec string) *openapi3.T {
	res, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoErrorf(t, err, "invalid OpenAPI spec")

	return res
}

type testCase struct {
	name    string
	options Options
	spec    string
	res     string
}

func TestAmbassador(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)
			spec := parseSpec(t, testCase.spec)

			mappings, err := GenerateMappings(testCase.options, spec)
			r.NoError(err)
			r.Equal(testCase.res, mappings)
		})
	}
}

var testCases = []testCase{
	{
		name: "basic",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
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
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-updatepet
  namespace: ambassador
spec:
  prefix: "/pet"
  method: PUT
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "basic+namespace",
		options: Options{
			AmbassadorNamespace: "amb",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
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
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-updatepet
  namespace: amb
spec:
  prefix: "/pet"
  method: PUT
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "parameter",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
paths:
  "/pet/{petId}/uploadImage":
    post:
      operationId: uploadFile
      parameters:
        - name: petId
          in: path
          description: ID of pet to update
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful operation
`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-uploadfile
  namespace: ambassador
spec:
  prefix: "/pet/([a-zA-Z0-9]*)/uploadImage"
  prefix_regex: true
  method: POST
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "empty-operationId",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
paths:
  "/pet/{petId}/uploadImage":
    post:
      parameters:
        - name: petId
          in: path
          description: ID of pet to update
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful operation
`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-postpetpetiduploadimage
  namespace: ambassador
spec:
  prefix: "/pet/([a-zA-Z0-9]*)/uploadImage"
  prefix_regex: true
  method: POST
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "basepath",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "/api/v3",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
paths:
  "/pet/{petId}/uploadImage":
    post:
      parameters:
        - name: petId
          in: path
          description: ID of pet to update
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful operation
`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-postpetpetiduploadimage
  namespace: ambassador
spec:
  prefix: "/api/v3/pet/([a-zA-Z0-9]*)/uploadImage"
  prefix_regex: true
  method: POST
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "basepath+rootonly",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "/api/v3",
			TrimPrefix:          "",
			RootOnly:            true,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
paths:
  "/pet":
    put:
      operationId: updatePet
      responses:
        '200':
          description: Successful operation
  "/pet/{petId}/uploadImage":
    post:
      operationId: uploadFile
      parameters:
        - name: petId
          in: path
          description: ID of pet to update
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful operation`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore
  namespace: ambassador
spec:
  prefix: "/api/v3"
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "basepath+trimprefix",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "/petstore/api/v3",
			TrimPrefix:          "/petstore",
			RootOnly:            false,
		},
		spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
paths:
  "/pet/{petId}/uploadImage":
    post:
      parameters:
        - name: petId
          in: path
          description: ID of pet to update
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful operation
`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-postpetpetiduploadimage
  namespace: ambassador
spec:
  prefix: "/petstore/api/v3/pet/([a-zA-Z0-9]*)/uploadImage"
  prefix_regex: true
  method: POST
  service: petstore.default
  regex_rewrite:
    pattern: '/petstore(.*)'
    substitution: '\1'
`,
	},
}
