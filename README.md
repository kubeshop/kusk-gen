# Kusk
<!-- Add buttons here -->

Developers running their apps in Kubernetes shouldn't have to worry about messing with resources that do not directly
relate to their applications.

Kusk (_driver in Swedish_) is **THE** tool that treats your openapi or swagger spec as a source of truth to automatically generate
supplementary resources for your Kubernetes cluster services in regard to mappings, monitors and alerts

Kusk handles the generation of these resources so developers don't have to.

# Demo-Preview
TODO
but for now

![Random GIF](https://media.giphy.com/media/ZVik7pBtu9dNS/giphy.gif)


# Table of contents

- [Kusk](#Kusk)
- [Demo-Preview](#demo-preview)
- [Table of contents](#table-of-contents)
- [Installation](#installation)
- [Usage](#usage)
  - [Ambassador Mappings](docs/ambassador.md)
  - [Linkerd Service Profiles](docs/linkerd.md)
  - [Nginx-Ingress Ingress Resources](docs/nginx-ingress.md)
- [Development](#development)
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

# Contribute
[(Back to top)](#table-of-contents)

Please refer to our organisation wide contributing guide for comprehensive instructions [here](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md)

# License
[(Back to top)](#table-of-contents)

[MIT](./LICENSE)
