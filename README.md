# Kusk
<!-- Add buttons here -->

Developers deploying their REST APIs in Kubernetes shouldn't have to worry about messing with resources that do not directly
relate to their applications.

Kusk (_coachman in Swedish_) is **THE** tool that treats your  OpenAPI or Swagger specification as a source of truth to generate
supplementary Kubernetes resources for your REST APIs in regard to mappings, security, monitoring, etc.

Kusk can even inspect what's installed in a cluster for tools it supports generating resources for and
automatically generate those resources based off your OpenAPI specification (check [kusk wizard](#wizard) ).

![kusk-overview](https://user-images.githubusercontent.com/14029650/129193622-b5f06b8d-845d-4b1e-adaf-34dd7b3e0108.png)


# Table of contents
- [Installation](#installation)
- [Usage](#usage)
  - [Ambassador Mappings](docs/ambassador.md)
  - [Linkerd Service Profiles](docs/linkerd.md)
  - [Nginx-Ingress Ingress Resources](docs/nginx-ingress.md)
- [OpenAPI extension](#openapi-extension)
- [Wizard](#wizard)
- [GitOps](#gitops)
- [Development](#development)
  - [Adding a generator](#adding-a-generator)
- [Contribute](#contribute)
- [License](#license)

# Installation
[(Back to top)](#table-of-contents)

#### Homebrew
TODO

#### Latest release on Github
`go install github.com/kubeshop/kusk@$VERSION`

#### From source
```shell
git clone git@github.com:kubeshop/kusk.git && \
cd kusk && \
go install
```

# Usage
[(Back to top)](#table-of-contents)

For more comprehensive instructions on individual generators, please refer to the dedicated document in the docs folder
for that generator.

For a run-through of what Kusk can do with the tools already installed in your cluster run:
`kusk wizard -i my-openapi-spec.yaml`

Or use one of our examples
`kusk wizard -i examples/booksapp/booksapp.yaml`

```shell
Usage:
  kusk [command]

Available Commands:
  ambassador    Generates Ambassador Mappings for your service
  completion    generate the autocompletion script for the specified shell
  help          Help about any command
  linkerd       Generates Linkerd Service Profiles for your service
  nginx-ingress Generates nginx-ingress resources
  wizard        Connects to current Kubernetes cluster and lists available generators

Flags:
  -h, --help   help for kusk

Use "kusk [command] --help" for more information about a command.
```

# OpenAPI extension
Kusk comes with an [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to accommodate everything within an OpenAPI spec to make that a real source of truth for all objects that can be generated. Every single CLI option can be set within the `x-kusk` extension, i.e. (`x-kusk` is at the spec's root):

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

# Wizard
Kusk comes with a `kusk wizard` interactive CLI to help you get started!
![wizard-gif](./docs/kusk-wizard.svg)

# GitOps
Kusk can be integrated within your GitOps environment to make your OpenAPI specification a real source of truth. As of now we support using kusk as an ArgoCD configuration management plugin. Please check the [guide](./docs/argocd.md)

# Development
[(Back to top)](#table-of-contents)

Checkout our Github actions for how we build and test the code [here](.github/workflows/go.yml)

Clone this repository and navigate inside the project folder and install the dependencies by running:
```shell
go get -d ./...
```

You can also just compile the project (and its dependencies) by running:
```shell
go build
```

or run the project directly:
```shell
go run main.go
```

To run the tests:
```shell
go test ./...
```

## Adding a generator
To add a generator one would need to implement [`generators.Interface`](./generators/interface.go) and register it's implementation by adding an element to [`generators.Registry`](./generators/generators.go). The CLI command would be constructed automatically and the parsed OpenAPI spec would be passed into the generator, along with path/method options extracted from `x-kusk` extension. The CLI options provided by the generator _must_ conform to the same naming scheme as JSON/YAML tags on options passed from `x-kusk` extension for automatic merge to work.

Check out [generators](./generators) folder and [Options](./options/options.go) for the examples.

# Contribute
[(Back to top)](#table-of-contents)

Please refer to our organisation wide contributing guide for comprehensive instructions [here](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md)

# License
[(Back to top)](#table-of-contents)

[MIT](./LICENSE)
