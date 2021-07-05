# kusk

kusk is a tool that treats your openapi or swagger spec as a source of truth to generate automatically
various custom resources for your Kubernetes cluster services in regards to mappings, monitors and alerts

We handle the generation of these resources so developers don't have to.

## Usage

```shell
Framework that makes an OpenAPI definition the source of truth for all API-related objects in a cluster (services, mappings, monitors, etc)

Usage:
  kusk [command]

Available Commands:
  ambassador  Generates ambassador mappings for your cluster from the provided api specification
  completion  generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
  -h, --help        help for kusk
  -i, --in string   file path to api spec file to generate mappings from. e.g. -in apispec.yaml

Use "kusk [command] --help" for more information about a command.
```

### Example
For a quick minimal example, run the following
```shell
./kusk -i examples/petstore.yaml ambassador --service-name petstore
```

## Development
Checkout our Github actions for how we build and test the code [here](.github/workflows/go.yml)

Clone this repository and navigate inside the project folder and install the dependencies by running:
```shell
go get -d ./...
```

You can also just compile the project (and its dependencies) by running:
```shell
go build
```

or run the project without compiling:
```shell
go run main.go
```

To run the tests:
```shell
go test ./...
```