package v2

import (
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/generators/ambassador"
	"github.com/kubeshop/kusk/options"
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
	generators.Registry["ambassador2"] = New()
}

func New() *Generator {
	return &Generator{
		abstractGenerator: ambassador.AbstractGenerator{
			MappingTemplate:   mappingTemplate,
			RateLimitTemplate: rateLimitTemplate,
		},
	}
}

type Generator struct {
	abstractGenerator ambassador.AbstractGenerator
}

func (g *Generator) ShortDescription() string {
	return "Generates Ambassador 2.0 Mappings for your service"
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Cmd() string {
	return "ambassador2"
}

func (g *Generator) Flags() *pflag.FlagSet {
	return g.abstractGenerator.Flags()
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	return g.abstractGenerator.Generate(opts, spec)
}
