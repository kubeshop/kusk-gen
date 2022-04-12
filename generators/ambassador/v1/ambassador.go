package v1

import (
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	"github.com/kubeshop/kusk-gen/generators"
	"github.com/kubeshop/kusk-gen/generators/ambassador"
	"github.com/kubeshop/kusk-gen/options"
)

var (
	mappingTemplate   *template.Template
	rateLimitTemplate *template.Template
)

func init() {
	mappingTemplate = template.New("mapping")
	mappingTemplate = template.Must(mappingTemplate.Parse(mappingTemplateRaw))

	rateLimitTemplate = template.New("rateLimit")
	rateLimitTemplate = template.Must(rateLimitTemplate.Parse(ambassador.RateLimitTemplateRaw))
}

func init() {
	generators.Registry["ambassador"] = New()
}

func New() *Generator {
	return &Generator{
		AbstractGenerator: ambassador.AbstractGenerator{
			MappingTemplate:   mappingTemplate,
			RateLimitTemplate: rateLimitTemplate,
		},
	}
}

type Generator struct {
	AbstractGenerator ambassador.AbstractGenerator
}

func (g *Generator) ShortDescription() string {
	return "Generates Ambassador Mappings for your service"
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Cmd() string {
	return "ambassador"
}

func (g *Generator) Flags() *pflag.FlagSet {
	return g.AbstractGenerator.Flags()
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	return g.AbstractGenerator.Generate(opts, spec)
}
