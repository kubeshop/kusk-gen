# Traefik V2

```bash
kusk-gen traefik

Usage:
  kusk traefik [flags]

Flags:
  -i, --in string                         file path to api spec file to generate mappings from. e.g. --in apispec.yaml
      --namespace string                  namespace for generated resources (default "default")
      --service.name string               target Service name
      --service.namespace string          namespace containing the target Service (default "default")
      --service.port int32                target Service port (default 80)
      --host string                       the Host header value to listen on
      --path.base string                  a base path for Service endpoints (default "/")
      --path.trim_prefix string           a prefix to trim from the URL before forwarding to the upstream Service
      --rate_limits.burst uint32          request per second burst
      --rate_limits.rps uint32            request per second rate limit
      --timeouts.idle_timeout uint32      idle connection timeout (seconds)
      --timeouts.request_timeout uint32   total request timeout (seconds)
  -h, --help
```

Traefik generator generates [Kubernetes IngressRoute](https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/) with related Middlewares 
and [ServersTransport](https://doc.traefik.io/traefik/routing/services/#serverstransport) resources for exposing HTTP and HTTPS routes from outside the cluster to services within the cluster.

All options that can be set via flags can also be set using our `x-kusk` OpenAPI extension in your specification.

CLI flags apply only at the global level i.e. applies to all paths and methods.

To override settings on the path or HTTP method level, you are required to use the x-kusk extension at that path in your API specification.

## Full Options Reference
| Name                         | CLI Option                     | OpenAPI Spec x-kusk label    | Descriptions                                                                                                       | Overwritable at path / method |
|------------------------------|--------------------------------|------------------------------|--------------------------------------------------------------------------------------------------------------------|-------------------------------|
| OpenAPI or Swagger File      | --in                           | N/A                          | Location of the OpenAPI or Swagger specification                                                                   | ❌                             |
| Namespace                    | --namespace                    | namespace                    | the namespace in which to create the generated resources (Required)                                                | ❌                             |
| Service Name                 | --service.name                 | service.name                 | the name of the service running in Kubernetes (Required)                                                           | ❌                             |
| Service Namespace            | --service.namespace            | service.namespace            | The namespace where the service named above resides (default value: default)                                       | ❌                             |
| Service Port                 | --service.port                 | service.port                 | Port the service is listening on (default value: 80)                                                               | ❌                             |
| Path Base                    | --path.base                    | path.base                    | Prefix for your resource routes                                                                                    | ❌                             |
| Path Trim Prefix             | --path.trim_prefix             | path.trim_prefix             | Trim the specified prefix from URl before passing request onto service                                             | ❌                             |
| Path split                   | --path.split                   | path.split                   | Boolean; whether or not to force generator to generate a mapping for each path                                     | ❌                             |
| Ingress Host                 | --host                         | host                         | The value to set the host field to in the Ingress resource                                                         | ❌                             |
| Rate limit (RPS)             | --rate_limits.rps              | rate_limits.rps              | Request per second rate limit                                                                                      | ✅                             |
| Rate limit (burst)           | --rate_limits.burst            | rate_limits.burst            | Rate limit burst                                                                                                   | ✅                             |
| Request Timeout              | --timeouts.request_timeout     | timeouts.request_timeout     | Total request timeout (seconds)                                                                                    | ✅                             |
| Idle Timeout                 | --timeouts.idle_timeout        | timeouts.idle_timeout        | Idle connection timeout (seconds)                                                                                  | ✅                             |
| CORS Origins                 | N/A                            | cors.origins                 | Array of origins                                                                                                   | ✅                             |
| CORS Methods                 | N/A                            | cors.methods                 | Array of methods                                                                                                   | ✅                             |
| CORS Headers                 | N/A                            | cors.headers                 | Array of headers                                                                                                   | ✅                             |
| CORS ExposeHeaders           | N/A                            | cors.expose_headers          | Array of headers to expose                                                                                         | ✅                             |
| CORS Credentials             | N/A                            | cors.credentials             | Boolean: enable credentials (default value: false)                                                                 | ✅                             |
| CORS Max Age                 | N/A                            | cors.max_age                 | Integer:how long the response to the preflight request can be cached for without sending another preflight request | ✅                             |

## Basic Usage

### CLI Flags

```shell
kusk-gen traefik -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: my-namespace
  service:
    name: webapp
    namespace:  my-service-namespace
    port: 7000
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: my-namespace
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
  name: webapp
  namespace: my-namespace
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("GET")
    services:
    - name: webapp
      namespace: my-namespace
      port: 7000
      serversTransport: webapp
```

## Base Path and Trim Prefix

Setting the Base path option allows your service to be identified with the base path acting as a prefix.

Setting the trim prefix options will create Traefik Middleware to trim the prefix before sending the
request onto the service.

### CLI Flags

```shell
kusk-gen traefik -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--path.base /my-app \
--path.trim_prefix /my-app
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: my-namespace
  service:
    name: webapp
    namespace: my-service-namespace
    port: 7000
  path:
    base: /my-app
    trim_prefix: /my-app
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: webapp-strip-prefix
  namespace: my-namespace
spec:
  stripPrefix:
    prefixes:
    - /my-app
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: my-namespace
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
  name: webapp
  namespace: my-namespace
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/my-app/") && Method("GET")
    middlewares:
    - name: webapp-strip-prefix
      namespace: my-namespace
    services:
    - name: webapp
      namespace: my-namespace
      port: 7000
      serversTransport: webapp
```

## Setting the Host

### CLI Flags

```shell
kusk-gen traefik -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--host mycustomhost.com
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: my-namespace
  service:
    name: webapp
    namespace:  my-service-namespace
    port: 7000
  host: mycustomhost.com
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: my-namespace
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
  name: webapp
  namespace: my-namespace
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: Host("mycustomhost.com") && PathPrefix("/") && Method("GET")
    services:
    - name: webapp
      namespace: my-namespace
      port: 7000
      serversTransport: webapp
```

## Setting timeouts

kusk-gen allows for setting a request timeout via flags or the x-kusk OpenAPI extension.

Traefik uses [ServersTransport](https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-serverstransport) CRD to control timeouts to backend service.

x-kusk option timeouts.request_timeout is used to set both responseHeaderTimeout and dialTimeout in CRD.

x-kusk option timeouts.idle_timeout is used to set idleConnTimeout that controls closing idle keep-alive connection to backend.

Zero (0) value of timeout in CRD means "No timeout".

### CLI Flags

```shell
kusk-gen traefik -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--timeouts.request_timeout 120
--timeouts.idle_timeout 120
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: my-namespace
  service:
    name: webapp
    namespace: my-service-namespace
    port: 7000
  timeouts:
    request_timeout: 120
    idle_timeout: 120
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: my-namespace
spec:
  forwardingTimeouts:
    dialTimeout: 120
    idleConnTimeout: 120
    responseHeaderTimeout: 120
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  creationTimestamp: null
  name: webapp
  namespace: my-namespace
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("GET")
    services:
    - name: webapp
      namespace: my-namespace
      port: 7000
      serversTransport: webapp
```

## Setting Rate Limits

kusk-gen allows for setting a Rate Limits via flags or the x-kusk OpenAPI extension.

Traefik uses [RateLimit Middleware](https://doc.traefik.io/traefik/middlewares/http/ratelimit/) for that.

x-kusk option rate_limits.rps is used to set rateLimit.average (requests per second) in CRD.

x-kusk option rate_limits.burst is used to set rateLimit.burst in CRD.

### CLI Flags

```shell
kusk-gen traefik -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--rate_limits.rps 20
--rate_limits.burst 100
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: my-namespace
  service:
    name: webapp
    namespace: my-service-namespace
    port: 7000
  rate_limits:
    rps: 20
    burst: 100
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: webapp-ratelimit
  namespace: my-namespace
spec:
  rateLimit:
    average: 20
    burst: 100
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: my-namespace
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
  name: webapp
  namespace: my-namespace
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("GET")
    middlewares:
    - name: webapp-ratelimit
      namespace: my-namespace
    services:
    - name: webapp
      namespace: my-namespace
      port: 7000
      serversTransport: webapp
```

## CORS

Via the x-kusk extension, you can set cors policies on your resources.

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
    port: 7000
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
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: webapp-cors
  namespace: booksapp
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
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: booksapp
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
  name: webapp
  namespace: booksapp
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("GET")
    middlewares:
    - name: webapp-cors
      namespace: booksapp
    services:
    - name: webapp
      namespace: booksapp
      port: 7000
      serversTransport: webapp
```

## Basic Path settings override

For this example, let's assume that one of the paths in the API specification should have different CORS headers than the rest.

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
    port: 7000
  cors:
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
  /:
    get: {}
  /books:
    x-kusk:
      cors:
        methods:
          - POST
        headers:
          - Other-Content-Type
        credentials: true
        expose_headers:
          - X-Other-Custom-Header
        max_age: 120
    post: {}
...
```

### Sample Output

```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: webapp-cors
  namespace: booksapp
spec:
  headers:
    accessControlAllowCredentials: true
    accessControlAllowHeaders:
    - Content-Type
    accessControlAllowMethods:
    - POST
    - GET
    - OPTIONS
    accessControlMaxAge: 86400
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  creationTimestamp: null
  name: webapp-books-cors
  namespace: booksapp
spec:
  headers:
    accessControlAllowCredentials: true
    accessControlAllowHeaders:
    - Other-Content-Type
    accessControlAllowMethods:
    - POST
    accessControlMaxAge: 120
---
apiVersion: traefik.containo.us/v1alpha1
kind: ServersTransport
metadata:
  creationTimestamp: null
  name: webapp
  namespace: booksapp
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
  name: webapp
  namespace: booksapp
spec:
  entryPoints:
  - web
  routes:
  - kind: Rule
    match: PathPrefix("/") && Method("GET")
    middlewares:
    - name: webapp-cors
      namespace: booksapp
    services:
    - name: webapp
      namespace: booksapp
      port: 7000
      serversTransport: webapp
  - kind: Rule
    match: PathPrefix("/books") && Method("POST")
    middlewares:
    - name: webapp-books-cors
      namespace: booksapp
    services:
    - name: webapp
      namespace: booksapp
      port: 7000
      serversTransport: webapp
```