package ambassador

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/spec"
)

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

			spec, err := spec.Parse([]byte(testCase.spec))
			r.NoError(err, "failed to parse spec")

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
		name: "basic-json",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
{
  "openapi": "3.0.2",
  "info": {
    "title": "Swagger Petstore - OpenAPI 3.0",
    "version": "1.0.5"
  },
  "paths": {
    "/pet": {
      "put": {
        "operationId": "updatePet",
        "responses": {
          "200": {
            "description": "Successful operation"
          }
        }
      }
    }
  }
}
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
		name: "basic-namespace",
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
		name: "basepath-rootonly",
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
		name: "basepath-trimprefix",
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
	{
		name: "swagger-yaml",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
swagger: "2.0"
info:
  version: 1.0.0
  title: Swagger Petstore
basePath: /v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      parameters:
        - name: limit
          in: query
          required: false
          type: integer
          format: int32
      responses:
        "200":
          description: A paged array of pets
          schema:
            $ref: '#/definitions/Pets'
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/Error'
    post:
      summary: Create a pet
      operationId: createPets
      responses:
        "201":
          description: Null response
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/Error'
  /pets/{petId}:
    get:
      operationId: showPetById
      parameters:
        - name: petId
          in: path
          required: true
          type: string
      responses:
        "200":
          description: Expected response to a valid request
          schema:
            $ref: '#/definitions/Pets'
        default:
          description: unexpected error
          schema:
            $ref: '#/definitions/Error'
definitions:
  Pet:
    type: "object"
    required:
      - id
      - name
    properties:
      id:
        type: integer
        format: int64
      name:
        type: string
      tag:
        type: string
  Pets:
    type: array
    items:
      $ref: '#/definitions/Pet'
  Error:
    type: "object"
    required:
      - code
      - message
    properties:
      code:
        type: integer
        format: int32
      message:
        type: string
`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-createpets
  namespace: ambassador
spec:
  prefix: "/pets"
  method: POST
  service: petstore.default
  rewrite: ""
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-listpets
  namespace: ambassador
spec:
  prefix: "/pets"
  method: GET
  service: petstore.default
  rewrite: ""
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-showpetbyid
  namespace: ambassador
spec:
  prefix: "/pets/([a-zA-Z0-9]*)"
  prefix_regex: true
  method: GET
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "swagger-json",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			BasePath:            "",
			TrimPrefix:          "",
			RootOnly:            false,
		},
		spec: `
{
  "swagger": "2.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore"
  },
  "basePath": "/v1",
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "parameters": [
          {
            "name": "limit",
            "in": "query",
            "required": false,
            "type": "integer",
            "format": "int32"
          }
        ],
        "responses": {
          "200": {
            "description": "A paged array of pets",
            "schema": {
              "$ref": "#/definitions/Pets"
            }
          },
          "default": {
            "description": "unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "post": {
        "summary": "Create a pet",
        "operationId": "createPets",
        "responses": {
          "201": {
            "description": "Null response"
          },
          "default": {
            "description": "unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    },
    "/pets/{petId}": {
      "get": {
        "operationId": "showPetById",
        "parameters": [
          {
            "name": "petId",
            "in": "path",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "description": "Expected response to a valid request",
            "schema": {
              "$ref": "#/definitions/Pets"
            }
          },
          "default": {
            "description": "unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "Pet": {
      "type": "object",
      "required": [
        "id",
        "name"
      ],
      "properties": {
        "id": {
          "type": "integer",
          "format": "int64"
        },
        "name": {
          "type": "string"
        },
        "tag": {
          "type": "string"
        }
      }
    },
    "Pets": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/Pet"
      }
    },
    "Error": {
      "type": "object",
      "required": [
        "code",
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "message": {
          "type": "string"
        }
      }
    }
  }
}
`,
		res: `
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-createpets
  namespace: ambassador
spec:
  prefix: "/pets"
  method: POST
  service: petstore.default
  rewrite: ""
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-listpets
  namespace: ambassador
spec:
  prefix: "/pets"
  method: GET
  service: petstore.default
  rewrite: ""
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: petstore-showpetbyid
  namespace: ambassador
spec:
  prefix: "/pets/([a-zA-Z0-9]*)"
  prefix_regex: true
  method: GET
  service: petstore.default
  rewrite: ""
`,
	},
	{
		name: "port specified",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			ServicePort:         443,
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
  service: petstore.default:443
  rewrite: ""
`,
	},
	{
		name: "port 0 specified",
		options: Options{
			AmbassadorNamespace: "",
			ServiceNamespace:    "default",
			ServiceName:         "petstore",
			ServicePort:         0,
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
}
