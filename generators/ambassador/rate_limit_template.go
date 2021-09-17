package ambassador

type rateLimitTemplateData struct {
	Group       string
	Name        string
	Operation   string
	Rate        uint32
	BurstFactor uint32
}

var rateLimitTemplateRaw = `{{range .}}
---
apiVersion: getambassador.io/v2
kind: RateLimit
metadata:
  name: {{.Name}}
spec:
  domain: ambassador
  limits:
    - pattern:
      - {{if .Group}}"generic_key": "kusk-group-{{.Group}}"{{else}}"generic_key": "kusk-operation-{{.Operation}}"{{end}}
        "remote-address": "*"
      rate: {{.Rate}}
      {{if .BurstFactor}}
      burstFactor: {{.BurstFactor}}
      {{end}}
      unit: second

{{end}}
`
