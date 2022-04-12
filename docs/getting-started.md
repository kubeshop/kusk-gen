#  Getting Started

## Installing Kusk

### Homebrew
`brew install kubeshop/kusk/kusk`

### Latest release on Github
`go install github.com/kubeshop/kusk-gen@$VERSION`

If you don't want to build it yourself, the [Releases](https://github.com/kubeshop/kusk-gen/releases) page contains already built binaries for all supported platforms.

Download it and unpack *kusk* to the directory of you choice.

### From source
```shell
git clone git@github.com:kubeshop/kusk.git && \
cd kusk && \
go install
```

## Usage

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
  ingress-nginx Generates ingress-nginx resources
  traefik       Generates Traefik resources
  wizard        Connects to current Kubernetes cluster and lists available generators

Flags:
  -h, --help   help for kusk

Use "kusk [command] --help" for more information about a command.
```

For more comprehensive instructions on individual generators, please refer to the dedicated document in the docs folder
for that generator.
