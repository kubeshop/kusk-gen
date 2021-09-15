# Example: Petstore

Sample petstore application from Swagger.

## Create cluster
```shell
k3d cluster create -p "8080:80@loadbalancer" -p "8443:443@loadbalancer" cl1
```

Apply Petstore manifest

```shell
kubectl apply -f examples/petstore/manifest.yaml
```

## Ambassador Mappings
### Setup

```shell
kubectl apply -f https://www.getambassador.io/yaml/aes-crds.yaml && \
kubectl wait --for condition=established --timeout=90s crd -lproduct=aes && \
kubectl apply -f https://www.getambassador.io/yaml/aes.yaml && \
kubectl -n ambassador wait --for condition=available --timeout=90s deploy -lproduct=aes
```

### Generate mappings and curl service

Root only
```shell
# This will allow you to resolve the swagger documentation in the browser at https://localhost:8443/
go run main.go ambassador -i examples/petstore/petstore.yaml --path.base="/petstore" --path.trim_prefix="/petstore" --service.name "petstore" | kubectl apply -f -
```

CQRS Pattern
```shell
# This will allow you to resolve the swagger documentation in the browser at https://localhost:8443/
go run main.go ambassador -i examples/petstore/petstore.yaml --path.base="/" --service.name "petstore" | kubectl apply -f -

# This will create mappings for each route in the api
go run main.go ambassador -i examples/petstore/petstore.yaml --path.base="/petstore/api/v3" --path.trim_prefix="/petstore" --service.name "petstore" --path.split=true | kubectl apply -f -
curl -kLi 'https://localhost:8443/petstore/api/v3/pet/findByStatus?status=available'  
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
go run main.go nginx-ingress -i examples/petstore/petstore.yaml --path.base="/" --service.name "petstore" | kubectl apply -f -

# cURL an api endpoint
curl -kLi 'http://localhost:8080/api/v3/pet/findByStatus?status=available'

# Remove ingress patch file
rm ingress.yaml
```

## Traefik V2

Traefik is installed into K3s cluster by default, no need to setup anything.

### Generate the ingress resource from the specification

```shell
go run main.go traefik -i examples/petstore/petstore.yaml --path.base="/api/v3" --service.name "petstore" | kubectl apply -f -
```

### cURL an api endpoint

```shell
curl -kLi 'http://localhost:8080/api/v3/pet/findByStatus?status=available'
```


## Cleanup

```shell
k3d cluster delete cl1
```
