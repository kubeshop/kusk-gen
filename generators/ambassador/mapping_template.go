package ambassador

type corsTemplateData struct {
	Origins        string
	Methods        string
	Headers        string
	ExposedHeaders string

	Credentials bool
	MaxAge      string
}

type mappingTemplateData struct {
	MappingName string

	MappingNamespace string
	ServiceURL       string

	BasePath   string
	TrimPrefix string

	Method  string
	Path    string
	Regex   bool
	Rewrite bool

	Host      string
	HostRegex bool

	CORSEnabled bool

	CORS corsTemplateData

	LabelsEnabled bool

	RequestTimeout uint32
	IdleTimeout    uint32

	RateLimitGroup string
}

var ambassadorMappingTemplateRaw = `{{range .}}
---
apiVersion: x.getambassador.io/v3alpha1
kind: AmbassadorMapping
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
