package ambassador

var mappingTemplateRaw = `{{range .}}
---
apiVersion: getambassador.io/v2
kind:  Mapping
metadata:
  name: {{.ServiceName}}-{{.OperationName}}
  namespace: {{.Namespace}}
spec:
  prefix: "{{.Path}}" {{if .Regex}}
  prefix_regex: true{{end}}
  method: {{.Method}}
  service: {{.ServiceName}}
{{end}}
`
