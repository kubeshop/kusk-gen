# Nginx Ingress

```bash
kusk nginx-ingress

Usage:
  kusk nginx-ingress [flags]

Flags:
  -i, --in string                             file path to api spec file to generate mappings from. e.g. --in apispec.yaml
      --namespace string                      namespace for generated resources (default "default")
      --service.name string                   target Service name
      --service.namespace string              namespace containing the target Service (default "default")
      --service.port int32                    target Service port (default 80)
      --host string                           an Ingress Host to listen on
      --nginx_ingress.rewrite_target string   a custom NGINX rewrite target
      --path.base string                      a base path for Service endpoints (default "/")
      --path.trim_prefix string               a prefix to trim from the URL before forwarding to the upstream Service
  -h, --help                                  help for nginx-ingress
```

The nginx-ingress generator generates [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource) 
resources for exposing HTTP and HTTPS routes from outside the cluster to services within the cluster. 

All options that can be set via flags can also be set using our `x-kusk` OpenAPI extension in your specification.

CLI flags apply only at the global level i.e. applies to all paths and methods.

To override settings on the path or HTTP method level, you are required to use the x-kusk extension at that path in your API specification.ß

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
| Nginx Ingress Rewrite Target | --nginx_ingress.rewrite_target | nginx_ingress.rewrite_target | Manually set the rewrite target for where traffic must be redirected                                               | ❌                             |
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
kusk nginx-ingress -i examples/booksapp/booksapp.yaml \
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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  creationTimestamp: null
  name: webapp-ingress
  namespace: my-namespace
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
```

## Split Path
By setting split path to true, kusk will generate an Ingress per route specified in the
provided OpenAPI specification

### CLI Flags
```shell
kusk nginx-ingress -i examples/booksapp/booksapp.yaml \
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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  creationTimestamp: null
  name: books
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /books
            pathType: Exact
status:
  loadBalancer: {}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  creationTimestamp: null
  name: webapp-ingress
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
```


## Base Path and Trim Prefix
Setting the Base path option allows your service to be identified with the base path acting as a prefix.

Setting the trim prefix options will instruct ambassador to trim the prefix before sending the
request onto the service.

### CLI Flags
```shell
kusk nginx-ingress -i examples/booksapp/booksapp.yaml \
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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$2
  creationTimestamp: null
  name: webapp-ingress
  namespace: my-namespace
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /my-app(/|$)(.*)
            pathType: Prefix
status:
  loadBalancer: {}
```


## Setting the rewrite target
### CLI Flags
```shell
kusk nginx-ingress -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--nginx_ingress.rewrite_target /sometarget
```

### OpenAPI Specification
```yaml
openapi: 3.0.1
x-kusk:
  namespace: my-namespace
  nginx_ingress:
    rewrite_target: /sometarget
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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /sometarget
  creationTimestamp: null
  name: webapp-ingress
  namespace: my-namespace
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
```

## Setting the Host
### CLI Flags
```shell
kusk nginx-ingress -i examples/booksapp/booksapp.yaml \
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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  creationTimestamp: null
  name: webapp-ingress
  namespace: my-namespace
spec:
  ingressClassName: nginx
  rules:
    - host: mycustomhost.com
      http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
```

## Setting timeouts
kusk allows for setting a request timeout via flags or the x-kusk OpenAPI extension

The nginx-ingress generator will spread the total request time over the following settings, diving it by 2
- `nginx.ingress.kubernetes.io/proxy-send-timeout`
- `nginx.ingress.kubernetes.io/proxy-read-timeout`

### CLI Flags
```shell
kusk nginx-ingress -i examples/booksapp/booksapp.yaml \
--namespace my-namespace \
--service.name webapp \
--service.port 7000 \
--service.namespace my-service-namespace \
--timeouts.request_timeout 120
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
paths:
  /:
    get: {}
...
```

### Sample Output
```yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
  creationTimestamp: null
  name: webapp-ingress
  namespace: my-namespace
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
```


## CORS
Via the x-kusk extension, you can set cors policies on your resources.

Due to a limitation of the nginx-ingress controller, we can only choose a single origin (the first one) 
from `cors.origins`. Kusk logs a warning informing you of this.

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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/cors-allow-headers: Content-Type
    nginx.ingress.kubernetes.io/cors-allow-methods: POST, GET, OPTIONS
    nginx.ingress.kubernetes.io/cors-allow-origin: http://foo.example
    nginx.ingress.kubernetes.io/cors-expose-headers: X-Custom-Header
    nginx.ingress.kubernetes.io/cors-max-age: "86400"
    nginx.ingress.kubernetes.io/enable-cors: "true"
  creationTimestamp: null
  name: webapp-ingress
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
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
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/cors-allow-headers: Other-Content-Type
    nginx.ingress.kubernetes.io/cors-allow-methods: POST
    nginx.ingress.kubernetes.io/cors-expose-headers: X-Other-Custom-Header
    nginx.ingress.kubernetes.io/cors-max-age: "120"
    nginx.ingress.kubernetes.io/enable-cors: "true"
  creationTimestamp: null
  name: books
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /books
            pathType: Exact
status:
  loadBalancer: {}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/cors-allow-headers: Content-Type
    nginx.ingress.kubernetes.io/cors-allow-methods: POST, GET, OPTIONS
    nginx.ingress.kubernetes.io/cors-expose-headers: X-Custom-Header
    nginx.ingress.kubernetes.io/cors-max-age: "86400"
    nginx.ingress.kubernetes.io/enable-cors: "true"
  creationTimestamp: null
  name: webapp-ingress
  namespace: booksapp
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - backend:
              service:
                name: webapp
                port:
                  number: 7000
            path: /
            pathType: Prefix
status:
  loadBalancer: {}
```