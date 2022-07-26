# kusk-gen - use OpenAPI to configure Kubernetes

**This project is now deprecated** - please check out [Kusk Gateway](https://github.com/kubeshop/kusk-gateway), which applies the same philosophy as kusk-gen but exclusively for Envoy.

## What is kusk-gen?

Developers deploying their REST APIs in Kubernetes shouldn't have to worry about managing resources that do not directly
relate to their applications or services.

kusk-gen (_coachman in Swedish_) treats your OpenAPI/Swagger definition as a source of truth for generating 
supplementary Kubernetes resources for your REST APIs in regard to mappings, security, traffic-control, monitoring, etc.

Read the [Introductory blog-post](https://medium.com/kubeshop-i/hello-kusk-openapi-for-kubernetes-19be94fc1e91) to get an overview.

![kusk-gen-overview](https://user-images.githubusercontent.com/14029650/129193622-b5f06b8d-845d-4b1e-adaf-34dd7b3e0108.png)

## Quick Start

### Homebrew
`brew install kubeshop/kusk/kusk-gen`

### Latest release on Github
`go install github.com/kubeshop/kusk-gen@$VERSION`

If you don't want to build it yourself, the [Releases](https://github.com/kubeshop/kusk-gen/releases) page contains already built binaries for all supported platforms.

Download it and unpack *kusk-gen* to the directory of you choice.

### From source
```shell
git clone git@github.com:kubeshop/kusk-gen.git && \
cd kusk-gen && \
go install
```

Read more at [Getting Started](https://kubeshop.github.io/kusk-gen/getting-started/)

## Why kusk-gen?

Using OpenAPI as the source-for-truth for client, servers, testing, documentation, etc. is a common approach when 
building microservice architectures with REST APis. Kusk extends this paradigm to also include Kubernetes configurations, 
allowing you to 
- Cut down on development time when deploying your REST APIs to your clusters
- Remove the need to learn tools-specific formats and configurations
- Easily switch between supported tools without having to learn new formats/configurations

## Features

- kusk-gen can inspect your cluster for the tools it supports and generate corresponding resources automatically.
- the Kusk [OpenAPI Extension](https://kubeshop.github.io/kusk/openapi-extension/) allows you to specify extended QoS and k8s related metadata which will be used
  to configure your cluster accordingly.
- kusk-gen plays nicely with both manual and automated/GitOps/CD workflows.
- The underlying architecture makes it straight-forward to extend kusk-gen with new generators

kusk-gen currently supports (click for configuration options)
- [Ambassador 1.x](https://kubeshop.github.io/kusk-gen/ambassador/)
- [Ambassador 2.0](https://kubeshop.github.io/kusk-gen/ambassador2/)
  - **Warning** This is a developer preview and should be treated as unstable
- [Linkerd](https://kubeshop.github.io/kusk-gen/linkerd/)
- [Ingress-Nginx](https://kubeshop.github.io/kusk-gen/ingress-nginx/)
  - This generator refers to the community ingress from [Kubernetes ingress-nginx](https://github.com/kubernetes/ingress-nginx/)
- [Traefik V2 (v2.x)](https://kubeshop.github.io/kusk-gen/traefik/)

Some of the upcoming tools we'd like to support are Kong and Contour. Please don't hesitate to 
suggest others or contribute your own generator!

## Documentation & Support

To learn more about kusk-gen check out the [complete documentation](https://kubeshop.github.io/kusk-gen/)

Join our [Discord Server](https://discord.gg/uNuhy6GDyn) to ask questions, suggest ideas, etc.

# How to contribute

- Check out our [Contributor Guide](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md) and
  [Code of Conduct](https://github.com/kubeshop/.github/blob/main/CODE_OF_CONDUCT.md)
- Fork/Clone the repo and make sure you can run it as shown above
- Check out open [issues](https://github.com/kubeshop/kusk-gen/issues) here on GitHub
- Get in touch with the team by starting a discussion on [GitHub](https://github.com/kubeshop/kusk-gen/discussions) or on our [Discord Server](https://discord.gg/uNuhy6GDyn).
  or open an issue of your own that you would like to contribute to the project.
- Fly like the wind!
