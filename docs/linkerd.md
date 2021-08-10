# Ambassador

```shell
kusk linkerd

Usage:
  kusk linkerd [flags]

Flags:
  -i, --in string                         file path to api spec file to generate mappings from. e.g. --in apispec.yaml
      --namespace string                  namespace for generated resources (default "default")
      --service.name string               target Service name
      --service.namespace string          namespace containing the target Service (default "default")
      --service.port int32                target Service port (default 80)
      --cluster.cluster_domain string     kubernetes cluster domain (default "cluster.local")
      --path.base string                  a base prefix for Service endpoints (default "/")
      --timeouts.request_timeout uint32   total request timeout (seconds)
  -h, --help                              help for linkerd
```

The Linkerd generator generates [Service Profile](https://linkerd.io/2.10/features/service-profiles/) resources to provide Linkerd information about routes to your service

All options that can be set via flags can also be set using our `x-kusk` OpenAPI extension in your specification.

CLI flags apply only at the global level i.e. applies to all paths and methods.

To override settings on the path or HTTP method level, you are required to use the x-kusk extension at that path in your API specification.

# Table of contents

- [Full Options Reference](#full-options-reference)
- [Basic Usage](#basic-usage)
- [Base Path](#base-path)
- [Change cluster domain](#change-cluster-domain)
- [Setting timeouts](#setting-timeouts)
- [Basic Path settings override](#basic-path-settings-override)

## Full Options Reference
|           Name          |         CLI Option         | OpenAPI Spec x-kusk label |                                 Descriptions                                 | Overwritable at path / method  |
|:-----------------------:|:--------------------------:|:-------------------------:|:----------------------------------------------------------------------------:|:------------------------------:|
| OpenAPI or Swagger File |            --in            |            N/A            |               Location of the OpenAPI or Swagger specification               |                ❌               |
|        Namespace        |         --namespace        |         namespace         |      the namespace in which to create the generated resources (Required)     |                ❌               |
|       Service Name      |       --service.name       |        service.name       |           the name of the service running in Kubernetes (Required)           |                ❌               |
|    Service Namespace    |     --service.namespace    |     service.namespace     | The namespace where the service named above resides (default value: default) |                ❌               |
|       Service Port      |       --service.port       |        service.port       |             Port the service is listening on (default value: 80)             |                ❌               |
|        Path Base        |         --path.base        |         path.base         |                        Prefix for your resource routes                       |                ❌               |
|      Cluster Domain     |  --cluster.cluster_domain  |   cluster.cluster_domain  |  Override the default internal cluster domain (default: cluster.local)       |                ❌               |
|     Request Timeout     | --timeouts.request_timeout |  timeouts.request_timeout |                        Total request timeout (seconds)                       |                ✅               |

## Basic Usage
### CLI Flags
```shell
kusk linkerd -i examples/booksapp/booksapp.yaml \
--service.name webapp \
--namespace my-namespace \
--service.namespace my-service-namespace
```

### OpenAPI Specification
```yaml
openapi: 3.0.1
x-kusk:
  namespace: booksapp
  service:
    name: webapp
    namespace: my-service-namespace
    port: 7000
paths:
  /:
    get: {}
...
```

### Sample Output
```yaml
---
apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.my-service-namespace.svc.cluster.local
  namespace: my-namespace
spec:
  routes:
    - condition:
        method: GET
        pathRegex: /
      name: GET /
    - condition:
        method: GET
        pathRegex: /authors/[^/]*
      name: GET /authors/{id}
    - condition:
        method: GET
        pathRegex: /books/[^/]*
      name: GET /books/{id}
    - condition:
        method: POST
        pathRegex: /authors
      name: POST /authors
    - condition:
        method: POST
        pathRegex: /authors/[^/]*/delete
      name: POST /authors/{id}/delete
    - condition:
        method: POST
        pathRegex: /authors/[^/]*/edit
      name: POST /authors/{id}/edit
    - condition:
        method: POST
        pathRegex: /books
      name: POST /books
    - condition:
        method: POST
        pathRegex: /books/[^/]*/delete
      name: POST /books/{id}/delete
    - condition:
        method: POST
        pathRegex: /books/[^/]*/edit
      name: POST /books/{id}/edit
```

## Base Path
Setting the Base path option allows your service to be identified with the base path acting as a prefix.

### CLI Flags
```shell
kusk linkerd -i examples/booksapp/booksapp.yaml \
--service.name webapp \
--namespace my-namespace \
--service.namespace my-service-namespace \
--path.base /my-app
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
paths:
  /:
    get: {}
...
```

### Sample Output
```yaml
---
apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.my-service-namespace.svc.cluster.local
  namespace: my-namespace
spec:
  routes:
    - condition:
        method: GET
        pathRegex: /my-app/
      name: GET /my-app/
    - condition:
        method: GET
        pathRegex: /my-app/authors/[^/]*
      name: GET /my-app/authors/{id}
    - condition:
        method: GET
        pathRegex: /my-app/books/[^/]*
      name: GET /my-app/books/{id}
    - condition:
        method: POST
        pathRegex: /my-app/authors
      name: POST /my-app/authors
    - condition:
        method: POST
        pathRegex: /my-app/authors/[^/]*/delete
      name: POST /my-app/authors/{id}/delete
    - condition:
        method: POST
        pathRegex: /my-app/authors/[^/]*/edit
      name: POST /my-app/authors/{id}/edit
    - condition:
        method: POST
        pathRegex: /my-app/books
      name: POST /my-app/books
    - condition:
        method: POST
        pathRegex: /my-app/books/[^/]*/delete
      name: POST /my-app/books/{id}/delete
    - condition:
        method: POST
        pathRegex: /my-app/books/[^/]*/edit
      name: POST /my-app/books/{id}/edit
```

## Change cluster domain
Setting the Base path option allows your service to be identified with the base path acting as a prefix.

### CLI Flags
```shell
kusk linkerd -i examples/booksapp/booksapp.yaml \
--service.name webapp \
--namespace my-namespace \
--service.namespace my-service-namespace \
--cluster.cluster_domain my-cluster.domain
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
  cluster:
    cluster_domain: my-cluster.domain
paths:
  /:
    get: {}
...
```

### Sample Output
```yaml
---
apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.my-service-namespace.svc.cluster.local
  namespace: my-namespace
spec:
  routes:
    - condition:
        method: GET
        pathRegex: /my-app/
      name: GET /my-app/
    - condition:
        method: GET
        pathRegex: /my-app/authors/[^/]*
      name: GET /my-app/authors/{id}
    - condition:
        method: GET
        pathRegex: /my-app/books/[^/]*
      name: GET /my-app/books/{id}
    - condition:
        method: POST
        pathRegex: /my-app/authors
      name: POST /my-app/authors
    - condition:
        method: POST
        pathRegex: /my-app/authors/[^/]*/delete
      name: POST /my-app/authors/{id}/delete
    - condition:
        method: POST
        pathRegex: /my-app/authors/[^/]*/edit
      name: POST /my-app/authors/{id}/edit
    - condition:
        method: POST
        pathRegex: /my-app/books
      name: POST /my-app/books
    - condition:
        method: POST
        pathRegex: /my-app/books/[^/]*/delete
      name: POST /my-app/books/{id}/delete
    - condition:
        method: POST
        pathRegex: /my-app/books/[^/]*/edit
      name: POST /my-app/books/{id}/edit
```


## Setting timeouts
kusk's Linkerd generator allows for setting request timeouts via flags or the x-kusk OpenAPI extension

### CLI Flags
```shell
kusk linkerd -i examples/booksapp/booksapp.yaml \
--service.name webapp \
--namespace my-namespace \
--service.namespace my-service-namespace \
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
paths:
  /:
    get: {}
...
```

### Sample Output
```yaml
---
apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.my-service-namespace.svc.cluster.local
  namespace: my-namespace
spec:
  routes:
    - condition:
        method: GET
        pathRegex: /
      name: GET /
      timeout: 120s
    - condition:
        method: GET
        pathRegex: /authors/[^/]*
      name: GET /authors/{id}
      timeout: 120s
    - condition:
        method: GET
        pathRegex: /books/[^/]*
      name: GET /books/{id}
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /authors
      name: POST /authors
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /authors/[^/]*/delete
      name: POST /authors/{id}/delete
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /authors/[^/]*/edit
      name: POST /authors/{id}/edit
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /books
      name: POST /books
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /books/[^/]*/delete
      name: POST /books/{id}/delete
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /books/[^/]*/edit
      name: POST /books/{id}/edit
      timeout: 120s
```

## Basic Path settings override
For this example, let's assume that one of the paths in the API specification should have a different request timeout value than the rest.

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
paths:
  /:
    get: {}
  /books:
    x-kusk:
      timeouts:
        request_timeout: 60
    post: {}
...
```

### Sample Output
```yaml
---
apiVersion: linkerd.io/v1alpha2
kind: ServiceProfile
metadata:
  creationTimestamp: null
  name: webapp.my-service-namespace.svc.cluster.local
  namespace: my-namespace
spec:
  routes:
    - condition:
        method: GET
        pathRegex: /
      name: GET /
      timeout: 120s
    - condition:
        method: POST
        pathRegex: /books
      name: POST /books
      timeout: 60s
```
