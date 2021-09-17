package traefik

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/structs"
	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/spec"
)

func TestTraefik(t *testing.T) {
	var testCases = []struct {
		name string
		spec string
		res  string
	}{
		{
			name: "root base path and no trim prefix",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    namespace: nondefault
    port: 7000
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
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/pet") && Method("PUT")
    services:
    - name: petstore
      namespace: nondefault
      port: 7000
      serversTransport: petstore
`,
		},
		{
			name: "non-root path and no trim prefix",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    namespace: nondefault
    port: 7000
  path:
    base: "/somepath"
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
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/somepath/pet") && Method("PUT")
    services:
    - name: petstore
      namespace: nondefault
      port: 7000
      serversTransport: petstore
`,
		},
		{
			name: "non-root path and trim prefix",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    namespace: nondefault
    port: 7000
  path:
    base: "/somepath"
    trim_prefix: "/somepath"
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
kind: Middleware
metadata:
  creationTimestamp: null
  name: petstore-strip-prefix
  namespace: nondefault
spec:
  stripPrefix:
    prefixes:
    - /somepath
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/somepath/pet") && Method("PUT")
    middlewares:
    - name: petstore-strip-prefix
      namespace: nondefault
    services:
    - name: petstore
      namespace: nondefault
      port: 7000
      serversTransport: petstore
`,
		},
		{
			name: "non-root path and host match",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  host: "example.com"
  service:
    name: petstore
    namespace: nondefault
    port: 7000
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
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: Host("example.com") && PathPrefix("/pet") && Method("PUT")
    services:
    - name: petstore
      namespace: nondefault
      port: 7000
      serversTransport: petstore
`,
		},
		{
			name: "root path with CORS, timeouts and rate limites overrides per path/method",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    port: 7777
    namespace: nondefault
  timeouts:
    request_timeout: 333
    idle_timeout: 22
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
    max_age: 86400
paths:
  "/pet":
    x-kusk:
      cors:
        origins:
          - http://foobar.example
        methods:
          - POST
          - GET
          - OPTIONS
        headers:
          - Content-Type
        credentials: true
        max_age: 12000
      timeouts:
        request_timeout: 20
        idle_timeout: 10
    put:
      x-kusk:
        disabled: true
      operationId: updatePet
      responses:
        '200':
          description: Successful operation
    post:
      operationId: addPet
      x-kusk:
        timeouts:
          request_timeout: 20
          idle_timeout: 10
        cors:
          origins:
            - http://putfoobar.example
          methods:
            - PUT
            - POST
            - OPTIONS
          headers:
            - Content-Type
          credentials: true
          max_age: 14000
  "/pet/findByStatus":
     get:
       x-kusk:
         cors:
           origins:
           - http://bar.example
           methods:
           - GET
           headers:
           - Content-Type
           credentials: false
           expose_headers:
           - X-Custom-Header
           max_age: 86400
         rate_limits:
           rps: 40
           burst: 80
`,
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: petstore-cors
  namespace: nondefault
spec:
  headers:
    accessControlAllowCredentials: true
    accessControlAllowHeaders:
    - Content-Type
    accessControlAllowMethods:
    - POST
    - GET
    - OPTIONS
    accessControlAllowOriginList:
    - http://foo.example
    - http://bar.example
    accessControlMaxAge: 86400
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: petstore-pet-cors
  namespace: nondefault
spec:
  headers:
    accessControlAllowCredentials: true
    accessControlAllowHeaders:
    - Content-Type
    accessControlAllowMethods:
    - POST
    - GET
    - OPTIONS
    accessControlAllowOriginList:
    - http://foobar.example
    accessControlMaxAge: 12000
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: petstore-pet-post-cors
  namespace: nondefault
spec:
  headers:
    accessControlAllowCredentials: true
    accessControlAllowHeaders:
    - Content-Type
    accessControlAllowMethods:
    - PUT
    - POST
    - OPTIONS
    accessControlAllowOriginList:
    - http://putfoobar.example
    accessControlMaxAge: 14000
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: petstore-petfindbystatus-get-cors
  namespace: nondefault
spec:
  headers:
    accessControlAllowHeaders:
    - Content-Type
    accessControlAllowMethods:
    - GET
    accessControlAllowOriginList:
    - http://bar.example
    accessControlMaxAge: 86400
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: petstore-petfindbystatus-ratelimit
  namespace: nondefault
spec:
  rateLimit:
    average: 40
    burst: 80
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 333
    idleConnTimeout: 22
    responseHeaderTimeout: 333
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore-pet
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 20
    idleConnTimeout: 10
    responseHeaderTimeout: 20
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore-pet-post
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 20
    idleConnTimeout: 10
    responseHeaderTimeout: 20
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/pet") && Method("POST")
    middlewares:
    - name: petstore-pet-post-cors
      namespace: nondefault
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore-pet-post
  - kind: Rule
    match: PathPrefix("/pet/findByStatus") && Method("GET")
    middlewares:
    - name: petstore-petfindbystatus-get-cors
      namespace: nondefault
    - name: petstore-petfindbystatus-ratelimit
      namespace: nondefault
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore
`,
		},
		{
			name: "globally disabled",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  disabled: true
  namespace: nondefault
  service:
    name: petstore
    port: 7777
    namespace: nondefault
paths:
  /:
    get: {}
    post: {}
`,
			res: ``,
		},
		{
			name: "path disabled, operation enabled",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    port: 7777
    namespace: nondefault
paths:
  /:
    x-kusk:
      disabled: true
    get:
      x-kusk:
        disabled: false
    post: {}
`,
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("GET")
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore
`,
		},
		{
			name: "path disabled not specified operation disabled specified",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    port: 7777
    namespace: nondefault
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
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("PATCH")
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore
  - kind: Rule
    match: PathPrefix("/") && Method("POST")
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore
`,
		},
		{
			name: "path disabled not specified operation disabled specified operation enabled not specified",
			spec: `
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
  version: 1.0.5
x-kusk:
  namespace: nondefault
  service:
    name: petstore
    port: 7777
    namespace: nondefault
paths:
  /:
    get:
      x-kusk:
        disabled: true
    post: {}
    patch: {}
`,
			res: `
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  forwardingTimeouts:
    dialTimeout: 0
    idleConnTimeout: 0
    responseHeaderTimeout: 0
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: petstore
  namespace: nondefault
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("PATCH")
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore
  - kind: Rule
    match: PathPrefix("/") && Method("POST")
    services:
    - name: petstore
      namespace: nondefault
      port: 7777
      serversTransport: petstore
`,
		},
	}

	var gen Generator

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			apiSpec, err := spec.NewParser(openapi3.NewLoader()).ParseFromReader(strings.NewReader(testCase.spec))
			r.NoError(err, "failed to parse spec")

			kuskExtensionOpts, err := spec.GetOptions(apiSpec)
			r.NoError(err, "failed to get options")
			k := koanf.New(".")
			err = k.Load(structs.Provider(*kuskExtensionOpts, "yaml"), nil)
			r.NoError(err)
			var opts options.Options
			err = k.UnmarshalWithConf("", &opts, koanf.UnmarshalConf{Tag: "yaml"})
			r.NoError(err)
			opts.PathSubOptions = kuskExtensionOpts.PathSubOptions
			opts.OperationSubOptions = kuskExtensionOpts.OperationSubOptions

			profile, err := gen.Generate(&opts, apiSpec)
			r.NoError(err)
			r.Equal(testCase.res, profile)
		})
	}
}
