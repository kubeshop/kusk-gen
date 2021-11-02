package v2

var mappingTemplateRaw = `{{range .}}
---
apiVersion: x.getambassador.io/v3alpha1
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
  hostname: '{{.Host}}'
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
    {{ $origins := split .CORS.Origins "," }}
    {{ if gt (len $origins) 0 }}
    origins:
    {{ range $val := $origins }}
      - "{{ $val }}"
    {{ end }}
    {{ end }}

    {{ $methods := split .CORS.Methods "," }}
    {{ if gt (len $methods) 0 }}
    methods:
    {{ range $val := $methods }}
      - "{{ $val }}"
    {{ end }}
    {{ end }}

    {{ $headers := split .CORS.Headers "," }}
    {{ if gt (len $methods) 0 }}
    headers:
    {{ range $val := $headers }}
      - "{{ $val }}"
    {{ end }}
    {{ end }}

    {{ $exposedHeaders := split .CORS.ExposedHeaders "," }}
    {{ if gt (len $exposedHeaders) 0 }}
    exposed_headers:
    {{ range $val := $exposedHeaders }}
      - "{{ $val }}"
    {{ end }}
    {{ end }}

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
