# For Developers

Checkout our Github actions for how we build and test the code [here](https://github.com/kubeshop/kusk/blob/main/.github/workflows/go.yml)

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

## Adding a custom generator

To add a generator for a tool not yet supported by Kusk one would need to implement [`generators.Interface`](https://github.com/kubeshop/kusk/blob/main/generators/interface.go)
and register it's implementation by adding an element to [`generators.Registry`](https://github.com/kubeshop/kusk/blob/main/generators/generators.go).
The CLI command would be constructed automatically and the parsed OpenAPI spec would be passed into the generator,
along with path/method options extracted from `x-kusk` extension. The CLI options provided by the generator _must_ conform to
the same naming scheme as JSON/YAML tags on options passed from `x-kusk` extension for automatic merge to work.

Check out [generators](https://github.com/kubeshop/kusk/blob/main/generators) folder and [Options](https://github.com/kubeshop/kusk/blob/main/options/options.go) for the examples.

##ß If you want to contribute

- Check out our [Contributor Guide](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md) and
  [Code of Conduct](https://github.com/kubeshop/.github/blob/main/CODE_OF_CONDUCT.md)
- Fork/Clone the repo and make sure you can run it as shown above
- Check out open [issues](https://github.com/kubeshop/monokle/issues) here on GitHub
- Get in touch with the team by starting a [discussion](https://github.com/kubeshop/kusk/discussions) on what you want to help with
  or open an issue of your own that you would like to contribute to the project.
- Fly like the wind!
