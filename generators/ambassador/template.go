package ambassador

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
}

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

{{end}}
`
