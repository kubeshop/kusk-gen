# Example: Petstore

Sample petstore application from Swagger.

## Prerequisites
This example uses k3d version v5.0.1 and Kubernetes version v1.22.2
```shell
❯ k3d --version
k3d version v5.0.1
k3s version latest (default)
```

```shell
kubectl version
Client Version: version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.4", GitCommit:"b695d79d4f967c403a96986f1750a35eb75e75f1", GitTreeState:"clean", BuildDate:"2021-11-17T15:41:42Z", GoVersion:"go1.16.10", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.2+k3s2", GitCommit:"3f5774b41eb475eb10c93bb0ce58459a6f777c5f", GitTreeState:"clean", BuildDate:"2021-10-05T20:29:33Z", GoVersion:"go1.16.8", Compiler:"gc", Platform:"linux/amd64"}
```

## Create cluster if not using Traefik
```shell
k3d cluster create cl1 \
-p 8080:80@loadbalancer \
-p 8443:443@loadbalancer \
--k3s-arg "--disable=traefik@server:0"
```
## Create cluster if using Traefik
```shell
k3d cluster create -p "8080:80@loadbalancer" -p "8443:443@loadbalancer" cl1
kubectl wait --for=condition=available --timeout=600s deployment/traefik -n kube-system
```

Apply Petstore manifest

```shell
kubectl apply -f examples/petstore/manifest.yaml
```

## Ambassador Mappings
### Setup

```shell
kubectl create namespace emissary && \
kubectl apply -f https://app.getambassador.io/yaml/emissary/latest/emissary-crds.yaml && \
kubectl wait --for condition=established --timeout=90s crd -lapp.kubernetes.io/name=ambassador && \
kubectl apply -f https://app.getambassador.io/yaml/emissary/latest/emissary-ingress.yaml && \
kubectl -n emissary wait --for condition=available --timeout=90s deploy -lproduct=aes
```

```shell
kubectl apply -f - <<EOF
---
apiVersion: getambassador.io/v3alpha1
kind: Listener
metadata:
  name: emissary-ingress-listener-8080
  namespace: emissary
spec:
  port: 8080
  protocol: HTTP
  securityModel: XFP
  hostBinding:
    namespace:
      from: ALL
EOF
```

### Generate mappings and curl service

Root only
```shell
# This will allow you to resolve the swagger documentation in the browser at https://localhost:8080/
go run main.go ambassador2 -i examples/petstore/petstore.yaml --path.base="/petstore" --path.trim_prefix="/petstore" --service.name "petstore" --host "*" | kubectl apply -f -

curl -Li 'http://localhost:8080/petstore/api/v3/pet/findByStatus?status=available'
```

CQRS Pattern
```shell
# This will allow you to resolve the swagger documentation in the browser at https://localhost:8080/
go run main.go ambassador2 -i examples/petstore/petstore.yaml --path.base="/" --service.name "petstore" --host "*" | kubectl apply -f -

# This will create mappings for each route in the api
go run main.go ambassador2 -i examples/petstore/petstore.yaml --path.base="/petstore/api/v3" --path.trim_prefix="/petstore" --service.name "petstore" --path.split=true --host "*" | kubectl apply -f -

curl -Li 'http://localhost:8080/petstore/api/v3/pet/findByStatus?status=available'
```

## Linkerd Service Profiles
### Setup
Install Linkerd using the following [guide](https://linkerd.io/2.10/getting-started/)

### Generate Service Profiles
```shell
go run main.go linkerd -i examples/petstore/petstore.yaml --path.base="/" --service.name "petstore" | kubectl apply -f -

# See metrics (if any)
linkerd viz routes svc/petstore

# See metrics for outgoing requests (if any)
linkerd viz routes deploy/petstore --to svc/petstore
```

## Nginx Ingress
### Setup
```yaml
# Install ingress-nginx
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/baremetal/deploy.yaml

# enable the ingress controller to use port 80 and 443 on the host
cat > ingress.yaml <<EOF
spec:
  template:
    spec:
      hostNetwork: true
EOF

# Patch the deployment
kubectl patch deployment ingress-nginx-controller -n ingress-nginx --patch "$(cat ingress.yaml)"

# Generate the ingress resource from the specification
go run main.go ingress-nginx -i examples/petstore/petstore.yaml --path.base="/" --service.name "petstore" | kubectl apply -f -

# cURL an api endpoint
curl -kLi 'http://localhost:8080/api/v3/pet/findByStatus?status=available'

# Remove ingress patch file
rm ingress.yaml
```

## Traefik V2

### Generate the ingress resource from the specification

```shell
go run main.go traefik -i examples/petstore/petstore.yaml --path.base="/api/v3" --service.name "petstore" | kubectl apply -f -
```

### cURL an api endpoint

```shell
curl -kLi 'http://localhost:8080/api/v3/pet/findByStatus?status=available'
```
It may take a few seconds for Traefik to resolve the IngressRoutes so retry the command if it fails

## Cleanup

```shell
k3d cluster delete cl1
```
