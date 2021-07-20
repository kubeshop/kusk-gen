package ambassador

type mappingTemplateData struct {
	ApiVersion  string
	Kind        string
	MappingName string

	AmbassadorNamespace string
	ServiceURL          string

	BasePath   string
	TrimPrefix string

	Hostname string
	Method   string
	Path     string
	Regex    bool
	Rewrite  bool
}

var mappingTemplateRaw = `{{range .}}
---
apiVersion: {{.ApiVersion}}
kind: {{.Kind}}
metadata:
  name: {{.MappingName}}
  namespace: {{.AmbassadorNamespace}}
spec:
  {{if .Hostname}}
  hostname: "{{.Hostname}}"
  {{end}}

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
