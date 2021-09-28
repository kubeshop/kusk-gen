# Ambassador

```shell
kusk ambassador2 --help
Generates Ambassador 2.0 Mappings for your service

Usage:
  kusk ambassador2 [flags]

Flags:
  -i, --in string                         file path to api spec file to generate mappings from. e.g. --in apispec.yaml
      --namespace string                  namespace for generated resources (default "default")
      --service.name string               target Service name
      --service.namespace string          namespace containing the target Service (default "default")
      --service.port int32                target Service port (default 80)
      --host string                       the Host header value to listen on
      --path.base string                  a base path for Service endpoints (default "/")
      --path.split                        force Kusk to generate a separate Mapping for each operation
      --path.trim_prefix string           a prefix to trim from the URL before forwarding to the upstream Service
      --rate_limits.burst uint32          request per second burst
      --rate_limits.rps uint32            request per second rate limit
      --timeouts.idle_timeout uint32      idle connection timeout (seconds)
      --timeouts.request_timeout uint32   total request timeout (seconds)
  -h, --help
```

The Ambassador generator generates [AmbassadorMapping]s (https://www.getambassador.io/docs/edge-stack/2.0/topics/using/intro-mappings/) resources for mapping resource to services. All options that can be set via 
flags can also be set using our `x-kusk` OpenAPI extension in your specification.

CLI flags apply only at the global level i.e. applies to all paths and methods.

To override settings on the path or HTTP method level, you are required to use the x-kusk extension at that path in your API specification.

## Full Options Reference
| Name                    | CLI Option                 | OpenAPI Spec x-kusk label | Descriptions                                                                                                       | Overwritable at path / method |
|-------------------------|----------------------------|---------------------------|--------------------------------------------------------------------------------------------------------------------|-------------------------------|
| OpenAPI or Swagger File | --in                       | N/A                       | Location of the OpenAPI or Swagger specification                                                                   | ❌                             |
| Namespace               | --namespace                | namespace                 | the namespace in which to create the generated resources (Required)                                                | ❌                             |
| Service Name            | --service.name             | service.name              | the name of the service running in Kubernetes (Required)                                                           | ❌                             |
| Service Namespace       | --service.namespace        | service.namespace         | The namespace where the service named above resides (default value: default)                                       | ❌                             |
| Service Port            | --service.port             | service.port              | Port the service is listening on (default value: 80)                                                               | ❌                             |
| Path Base               | --path.base                | path.base                 | Prefix for your resource routes                                                                                    | ❌                             |
| Path Trim Prefix        | --path.trim_prefix         | path.trim_prefix          | Trim the specified prefix from URl before passing request onto service                                             | ❌                             |
| Path split              | --path.split               | path.split                | Boolean; whether or not to force generator to generate a mapping for each path                                     | ❌                             |
| Host                    | --host                     | host                      | The value to set the host field to in the AmbassadorMapping resource                                                         | ✅                             |
| Rate limit (RPS)        | --rate_limits.rps          | rate_limits.rps           | Request per second rate limit                                                                                      | ✅                             |
| Rate limit (burst)      | --rate_limits.burst        | rate_limits.burst         | Rate limit burst                                                                                                   | ✅                             |
| Rate limit group        | N/A                        | rate_limits.group         | Rate limit endpoint group                                                                                          |                               |
| Request Timeout         | --timeouts.request_timeout | timeouts.request_timeout  | Total request timeout (seconds)                                                                                    | ✅                             |
| Idle Timeout            | --timeouts.idle_timeout    | timeouts.idle_timeout     | Idle connection timeout (seconds)                                                                                  | ✅                             |
| CORS Origins            | N/A                        | cors.origins              | Array of origins                                                                                                   | ✅                             |
| CORS Methods            | N/A                        | cors.methods              | Array of methods                                                                                                   | ✅                             |
| CORS Headers            | N/A                        | cors.headers              | Array of headers                                                                                                   | ✅                             |
| CORS ExposeHeaders      | N/A                        | cors.expose_headers       | Array of headers to expose                                                                                         | ✅                             |
| CORS Credentials        | N/A                        | cors.credentials          | Boolean: enable credentials (default value: false)                                                                 | ✅                             |
| CORS Max Age            | N/A                        | cors.max_age              | Integer:how long the response to the preflight request can be cached for without sending another preflight request | ✅                             |

## Ambassador 2.0 Setup
[source](https://www.getambassador.io/docs/edge-stack/latest/tutorials/getting-started/)

### Create cluster
`k3d cluster create -p "8080:80@loadbalancer" -p "8443:443@loadbalancer" --k3s-server-arg '--disable=traefik' cl1`

### Install the edge-stack
```
# Add the Repo:
helm repo add datawire https://www.getambassador.io
helm repo update

# Create Namespace and Install:
helm install -n ambassador --create-namespace \
        edge-stack --devel \
        datawire/edge-stack && \
    kubectl rollout status  -n ambassador deployment/edge-stack -w
```

### Create the AmbassadorListeners
```
kubectl apply -f - <<EOF
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorListener
metadata:
  name: edge-stack-listener-8080
  namespace: ambassador
spec:
  port: 8080
  protocol: HTTP
  securityModel: XFP
  hostBinding:
    namespace:
      from: ALL
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorListener
metadata:
  name: edge-stack-listener-8443
  namespace: ambassador
spec:
  port: 8443
  protocol: HTTPS
  securityModel: XFP
  hostBinding:
    namespace:
      from: ALL
EOF
```

### Create the AmbassadorHost
```
cat <<EOF | kubectl apply -f -
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorHost
metadata:
  name: my-host
  namespace: ambassador
spec:
  hostname: "*"
  requestPolicy:
    insecure:
      action: Redirect
EOF
```

### Deploy the example booksapp
```
kubectl create ns booksapp && \
kubectl apply -n booksapp -f /Users/kylehodgetts/kubeshop/kusk/examples/booksapp/manifest.yaml && \
kubectl rollout status  -n booksapp deployment/webapp -w
```

## Basic Usage

### CLI Flags

```shell
kusk ambassador2 -i examples/booksapp/booksapp.yaml \
--namespace booksapp \
--service.name webapp \
--service.port 7000 \
--service.namespace booksapp
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
    port: 7000
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: booksapp
  namespace: my-namespace
spec:
  prefix: "/"
  service: booksapp.my-service-namespace:7000
  rewrite: ""
```

## Split Path

By setting split path to true, kusk will generate a mapping per route specified in the 
provided OpenAPI specification

### CLI Flags

```shell
kusk ambassador -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name booksapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--path.split true
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
    port: 7000
  path:
    split: true
paths:
  /:
    get: {}

  /books:
    post: {}
...
```

### Sample Output

```yaml
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: booksapp-get
  namespace: my-namespace
spec:
  prefix: "/"
  method: GET
  service: booksapp.my-service-namespace:7000
  rewrite: ""
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: booksapp-postbooks
  namespace: my-namespace
spec:
  prefix: "/books"
  method: POST
  service: booksapp.my-service-namespace:7000
  rewrite: ""
```

## Base Path and Trim Prefix

Setting the Base path option allows your service to be identified with the base path acting as a prefix.

Setting the trim prefix options will instruct ambassador to trim the prefix before sending the 
request onto the service.

### CLI Flags

```shell
kusk ambassador -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name booksapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--path.base /my-app \
--path.trim_prefix /my-app
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
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
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: booksapp
  namespace: my-namespace
spec:
  prefix: "/my-app"
  service: booksapp.my-service-namespace:7000
  regex_rewrite:
    pattern: '/my-app(.*)'
    substitution: '\1'
```

## Setting timeouts

kusk allows for setting both idle and request timeouts via flags or the x-kusk OpenAPI extension

### CLI Flags

```shell
kusk ambassador -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name booksapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--timeouts.idle_timeout 120 \
--timeouts.request_timeout 120
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
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
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: booksapp
  namespace: my-namespace
spec:
  prefix: "/"
  service: booksapp.my-service-namespace:7000
  rewrite: ""
  timeout_ms: 120000
  idle_timeout_ms: 120000
```

## Setting the Host header

kusk allows for setting both idle and request timeouts via flags or the x-kusk OpenAPI extension

### CLI Flags

```shell
kusk ambassador -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name booksapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--host "somehost.io"
```

### OpenAPI Specification

```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: booksapp
    port: 7000
  host: somehost.io
paths:
  /:
    get: {}
...
```

### Sample Output

```yaml
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: booksapp
  namespace: my-namespace
spec:
  prefix: "/"
  host: somehost.io
  service: booksapp.my-service-namespace:7000
  rewrite: ""
  timeout_ms: 120000
  idle_timeout_ms: 120000
```

## CORS

Via the x-kusk extension, you can set cors policies on your resources

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
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: webapp
  namespace: booksapp
spec:
  prefix: "/"
  service: webapp.booksapp:7000
  rewrite: ""
  cors:
    origins: http://foo.example,http://bar.example
    methods: POST,GET,OPTIONS
    headers: Content-Type
    exposed_headers: X-Custom-Header
    credentials: true
    max_age: "86400"
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
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: webapp-get
  namespace: booksapp
spec:
  prefix: "/"
  method: GET
  service: webapp.booksapp:7000
  rewrite: ""
  cors:
    origins:
    methods: POST,GET,OPTIONS
    headers: Content-Type
    exposed_headers: X-Custom-Header
    credentials: true
    max_age: "86400"
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
metadata:
  name: webapp-postbooks
  namespace: booksapp
spec:
  prefix: "/books"
  method: POST
  service: webapp.booksapp:7000
  rewrite: ""
  cors:
    origins:
    methods: POST
    headers: Other-Content-Type
    exposed_headers: X-Other-Custom-Header
    credentials: true
    max_age: "120"
```
