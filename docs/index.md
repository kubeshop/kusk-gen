# Kusk
<!-- Add buttons here -->

Developers deploying their REST APIs in Kubernetes shouldn't have to worry about managing resources that do not directly
relate to their applications or services.

Kusk (_coachman in Swedish_) treats your OpenAPI/Swagger definition as a source of truth for generating
supplementary Kubernetes resources for your REST APIs in regard to mappings, security, traffic-control, monitoring, etc.

- The [Kusk wizard](#kusk-wizard) can inspect your cluster for the tools it supports and generate corresponding
  resources automatically.
- the [Kusk OpenAPI extension](openapi-extension.md) allows you to specify QoS and k8s related metadata which will be used
  to configure your cluster accordingly.
- Kusk plays nicely with both manual and automated/GitOps/CD workflows (see examples below).
- The underlying architecture makes it straight-forward to add new generators (see [Development](development.md))

![kusk-overview](https://user-images.githubusercontent.com/14029650/129193622-b5f06b8d-845d-4b1e-adaf-34dd7b3e0108.png)

## Kusk Wizard

Kusk comes with a `kusk wizard` interactive CLI to help you get started!
![wizard-gif](./kusk-wizard.svg)
