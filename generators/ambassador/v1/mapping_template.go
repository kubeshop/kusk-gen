package v1

var mappingTemplateRaw = `{{range .}}
---
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: {{.MappingName}}
  namespace: {{.MappingNamespace}}
spec:
  prefix: "{{.BasePath}}{{.Path}}" 

  {{if .Regex}}
  prefix_regex: true
  {{end}}

  {{ if .Host}}
  host: {{.Host}}
  {{end}}

  {{if .Method}}
  method: {{.Method}}
  {{end}}

  service: {{.ServiceURL}}

  {{if .TrimPrefix}}
  regex_rewrite:
    pattern: '{{.TrimPrefix}}(.*)'
    substitution: '\1'
  {{else}}
  rewrite: ""
  {{end}}

  {{if .CORSEnabled}}
  cors:
    origins: {{.CORS.Origins}}
    methods: {{.CORS.Methods}}
    headers: {{.CORS.Headers}}
    exposed_headers: {{.CORS.ExposedHeaders}}
    credentials: {{.CORS.Credentials}}
    max_age: "{{.CORS.MaxAge}}"
  {{end}}

  {{if .LabelsEnabled}}
  labels:
    ambassador:{{if .RateLimitGroup}}
	  - group:
		  - kusk-group-{{.RateLimitGroup}}{{else}}
      - operation:
          - kusk-operation-{{.MappingName}}{{end}}
      - request:
          - remote-address
  {{end}}

  {{if .RequestTimeout}}
  timeout_ms: {{.RequestTimeout}}
  {{end}}

  {{if .IdleTimeout}}
  idle_timeout_ms: {{.IdleTimeout}}
  {{end}}

{{end}}
`
