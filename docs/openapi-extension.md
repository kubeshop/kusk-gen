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
Please review the generator's documentation to see what can be overwritten.
