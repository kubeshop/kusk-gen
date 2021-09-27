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

	Host      string
	HostRegex bool

	CORSEnabled bool

	CORS corsTemplateData

	LabelsEnabled bool

	RequestTimeout uint32
	IdleTimeout    uint32

	RateLimitGroup string
}
