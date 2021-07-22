# Example: Petstore

Sample petstore application from Swagger.

## Create cluster
```shell
k3d cluster create -p "8080:80@loadbalancer" -p "8443:443@loadbalancer" --k3s-server-arg '--disable=traefik' cl1
```

Apply Petstore manifest
`kubectl apply -f examples/petstore/manifest.yaml`

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
go run main.go ambassador -i examples/petstore/petstore.yaml --base-path="/petstore" --trim-prefix="/petstore" --service-name "petstore" | kubectl apply -f -
```

CQRS Pattern
```shell
# This will allow you to resolve the swagger documentation in the browser at https://localhost:8443/
go run main.go ambassador -i examples/petstore/petstore.yaml --base-path="/" --service-name "petstore" | kubectl apply -f -

# This will create mappings for each route in the api
go run main.go ambassador -i examples/petstore/petstore.yaml --base-path="/petstore/api/v3" --trim-prefix="/petstore" --service-name "petstore" --root-only=false | kubectl apply -f -
curl -kLi 'https://localhost:8443/petstore/api/v3/pet/findByStatus?status=available'  
```

## Linkerd Service Profiles
[TODO]

## Nginx Ingress
[TODO]

## Cleanup
```shell
k3d cluster delete cl1
```
