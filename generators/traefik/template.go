package traefik

type ingressRouteData struct {
	Name             string
	Namespace        string
	Match            string
	ServiceName      string
	ServiceNamespace string
	ServicePort      int32
}

var ingressRouteTpl = `{{- range .}}
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  entryPoints:
    - web
  routes:
    - match: {{.Match}}
      kind: Rule
      services:
        - name: {{.ServiceName}}
          namespace: {{.ServiceNamespace}}
          port: {{.ServicePort}}
{{- end}}
`

var middlewareTpl = `{{range .}}
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: stripprefix
spec:
  stripPrefix:
    prefixes:
      - /stripit
{{end}}
`
