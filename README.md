# Kusk - use OpenAPI to configure Kubernetes 

## What is Kusk?

Developers deploying their REST APIs in Kubernetes shouldn't have to worry about managing resources that do not directly
relate to their applications or services.

Kusk (_coachman in Swedish_) treats your OpenAPI/Swagger definition as a source of truth for generating 
supplementary Kubernetes resources for your REST APIs in regard to mappings, security, traffic-control, monitoring, etc.

![kusk-overview](https://user-images.githubusercontent.com/14029650/129193622-b5f06b8d-845d-4b1e-adaf-34dd7b3e0108.png)

## Quick Start

### Latest release on Github
`go install github.com/kubeshop/kusk@$VERSION`

### From source
```shell
git clone git@github.com:kubeshop/kusk.git && \
cd kusk && \
go install
```

Read more at [Getting Started](https://kubeshop.github.io/kusk/getting-started/)

## Why Kusk?

Using OpenAPI as the source-for-truth for client, servers, testing, documentation, etc. is a common approach when 
building microservice architectures with REST APis. Kusk extends this paradigm to also include Kubernetes configurations, 
allowing you to 
- Cut down on development time when deploying your REST APIs to your clusters
- Remove the need to learn tools-specific formats and configurations
- Easily switch between supported tools without having to learn new formats/configurations

## Features

- Kusk can inspect your cluster for the tools it supports and generate corresponding resources automatically.
- the Kusk [OpenAPI Extension](https://kubeshop.github.io/kusk/openapi-extension/) allows you to specify extended QoS and k8s related metadata which will be used
  to configure your cluster accordingly.
- Kusk plays nicely with both manual and automated/GitOps/CD workflows.
- The underlying architecture makes it straight-forward to extend Kusk with new generators

Kusk currently supports
- Ambassador 1.x
- Linkerd
- Nginx-Ingress

Some of the upcoming tools we'd like to support are Traefik, Kong, Ambassador 2.x, etc. Please don't hesitate to 
suggest others or contribute your own generator!

## Documentation

To learn more about Kusk check out the [complete documentation](https://kubeshop.github.io/kusk/)

# How to contribute

- Check out our [Contributor Guide](https://github.com/kubeshop/.github/blob/main/CONTRIBUTING.md) and
  [Code of Conduct](https://github.com/kubeshop/.github/blob/main/CODE_OF_CONDUCT.md)
- Fork/Clone the repo and make sure you can run it as shown above
- Check out open [issues](https://github.com/kubeshop/kusk/issues) here on GitHub
- Get in touch with the team by starting a [discussion](https://github.com/kubeshop/kusk/discussions) on what you want to help with 
  or open an issue of your own that you would like to contribute to the project.
- Fly like the wind!
