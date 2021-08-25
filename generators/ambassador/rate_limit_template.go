package ambassador

type rateLimitTemplateData struct {
	Operation   string
	Rate        uint32
	BurstFactor uint32
}

var rateLimitTemplateRaw = `{{range .}}
---
apiVersion: getambassador.io/v2
kind: RateLimit
metadata:
  name: {{.Operation}}
spec:
  domain: ambassador
  limits:
    - pattern:
      - "generic_key": "kusk-operation-{{.Operation}}"
        "remote-address": "*"
      rate: {{.Rate}}
      {{if .BurstFactor}}
      burstFactor: {{.BurstFactor}}
      {{end}}
      unit: second

{{end}}
`
