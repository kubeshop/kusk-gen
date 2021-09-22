# OpenAPI extension
Kusk comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within 
an OpenAPI spec to make that a real source of truth for all objects that can be generated. Every single CLI option can be set 
within the `x-kusk` extension, i.e. (`x-kusk` is at the spec's root):

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
```
And more to that, `x-kusk` extension can also be used to overwrite specific options at the path/operation level, i.e.:

```yaml
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
```

`x-kusk` extension at the Operation level takes precedence, i.e. overrides, what's specified at the path level, including the `disabled` option.
Likewise, the Path level settings override what's specified at the global level.

If settings aren't specified at a path or operation level, it will inherit from the layer above. (Operation > Path > Global)

Please review the generator's documentation to see what can be overwritten.

# Merging vanilla OpenAPI yaml file and x-kusk extention

There are situations when you want to keep your OpenAPI file pristine and not add `x-kusk` extention to it.
E.g. if you generate it during the build each time, or if you have multiple environments that have different x-kusk overrides per env.

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
