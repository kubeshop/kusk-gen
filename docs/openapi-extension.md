# OpenAPI extension

Kusk comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within 
an OpenAPI spec to make that a real source of truth for all objects that can be generated. Every single CLI option can be set 
within the `x-kusk` extension. The extension can be specified at the root, path and operation levels.

## Properties Overview 

The following top-level properties are available:

| property | root | path | operation | [Amb 1.X](ambassador.md) | [Amb 2.X](ambassador2.md) | [LinkerD](linkerd.md) | [Nginx-Ing](ingress-nginx.md) | [Traefik](traefik.md)
| --- | :---: | :---: | :---: | :---: |  :---: |  :---: |  :---: |  :---: |   
| [`disabled`](#disabled) | X | X | X | X | X | X | X | X  
| [`host`](#host) | X | X | X | X | X | X | X | X
| [`cors`](#cors) | X | X | X | X | X |  | X | X
| [`rate_limits`](#rate-limits) | X | X | X |  | X | | X | X
| [`timeouts`](#timeouts) | X | X | X |  X | X | X | X | X
| [`namespace`](#namespace) | X |  |  |  X | X | X | X | X
| [`service`](#service) | X |  |  |  X | X | X | X | X
| [`path`](#path) | X |  |  |  X | X | X | X | X
| [`cluster`](#cluster) | X |  |  |   |  | X |  | 
| [`host`](#host) | X |  |  |  | X |  | X | X
| [`nginx_ingress`](#ingress-nginx) | X |  |  |  |  |  | X | 

### Property Overriding/inheritance

`x-kusk` extension at the operation level takes precedence, i.e. overrides, what's specified at the path level, including the `disabled` option.
Likewise, the path level settings override what's specified at the global level.

If settings aren't specified at a path or operation level, it will inherit from the layer above. (Operation > Path > Global)

## Top-level properties

### Disabled 

This boolean property allows you to disable the corresponding path/operation, allowing you to "hide" internal operations
from being published to end users. 

When set to true at the top level all paths will be hidden; you will have to override specific paths/operations with
`disabled: false` to make those operations visible.

### Host

This string property sets a corresponding [Ingress host rule](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules).

### CORS

The cors object sets properties for configuring [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for your API

| Name | Description |
| :---: | :--- |
| `origins` | list of HTTP origins accepted by the configured operations
| `methods` | list of HTTP methods accepted by the configured operations
| `headers` | list of HTTP headers accepted by the configured operations
| `expose_headers` | list of HTTP headers exposed by the configured operations
| `credentials` | boolean flag for requiring credentials
| `max_age` | the max age of the 

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### Rate Limits

Options for configuring rate-limiting.

| Name | Description |
| :---: | :--- |
| `rps` | requests-per-seconds
| `burst` | burst allowance
| `group` | rate-limiting group

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### Timeouts

Options for configuring request timeouts

| Name | Description |
| :---: | :--- |
| `request_timeout` | total request timeout (in seconds)
| `idle_timeout` | timeout for idle connections (in seconds)

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### Namespace

This string property sets the namespace for the generated resource. Default value is "default".

### Service

The service object sets the target service to receive traffic, it contains the following properties:

| Name | Description |
| :---: | :--- |
| `namespace` | the namespace containing the upstream Service
| `name` | the upstream Service's name
| `port` | the upstream Service's port. Default value is 80

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### Path

The path object contains the following properties to configure service endpoints paths:

| Name | Description |
| :---: | :--- |
| `base` | Base is the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route). Default value is "/"
| `trim_prefix` | TrimPrefix is the prefix that would be omitted from the URL when request is being forwarded to the upstream service, i.e. given that Base is set to "/petstore/api/v3", TrimPrefix is set to "/petstore", path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive is "/api/v3/pets".
| `split` | forces Kusk to generate a separate resource for each Path or Operation, where appropriate

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### Cluster

The cluster object contains a set of cluster-wide properties.

| Name | Description |
| :---: | :--- |
| `cluster_domain` | the base DNS domain for the cluster. Default value is "cluster.local".

Please see the documentation for each individual generator to see which of these properties they support and how they apply.

### Host

A string specifying an Ingress host rule - see 
https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules for additional documentation.

### Nginx Ingress

Options specific to the [ingress-nginx controller](ingress-nginx.md)

| Name | Description |
| :---: | :--- |
| `rewrite_target` | RewriteTarget is a custom rewrite target for ingress-nginx, see https://kubernetes.github.io/ingress-nginx/examples/rewrite/ for additional documentation.

## Basic Example

The following sets cors, service and path properties at the global level, but disables the PUT operation at /pet

```yaml
x-kusk:
  cors:
    origins:
      - http://foo.example
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
  service:
    name: petstore
    port: 80
  path:
    base: /petstore/api/v3
    trim_prefix: /petstore
  paths:
    "/pet":
      put:
        x-kusk:
          disabled: true
        tags:
          - pet
        summary: Update an existing pet
        description: Update an existing pet by Id
        operationId: updatePet
    ...
```

## Merging vanilla OpenAPI yaml file and x-kusk extension

There are situations when you want to keep your OpenAPI file pristine and not add `x-kusk` extension to it.
E.g. if you generate it during the build each time, or if you have multiple environments that have 
different x-kusk overrides per env.

You can always add `x-kusks` enabled YAML file with extention keys and merge it with your OpenAPI file.
The resulting file can be consumed by Kusk for Ingress generation.

For that, you'll need to use [yt](https://mikefarah.gitbook.io/yq) tool.

E.g.

*petstore.yaml*

```yaml
paths:
  "/pet":
    put:
      ...
    post:
      ...
```

and *x-kusk.yaml* enabled:

```yaml
x-kusk:
  disabled: false
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
  service:
    name: petstore
    port: 80
  path:
    base: /petstore/api/v3
    trim_prefix: /petstore
paths:
  "/pet":
    x-kusk:
      disabled: false
    put:
      x-kusk:
        disabled: true
```

Running the tool:

```shell
yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)'  x-kusks.yaml petstore.yaml
```
will produce merged yaml, ready to be used by Kusk. Note, though, that the order of keys in the resulting map can be different.

If you use JSON for your OpenAPI file, the same result can be achieved with [jq](https://stedolan.github.io/jq/).
