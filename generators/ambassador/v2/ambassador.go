package v2

import (
	"errors"
	"strings"
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
	mappingTemplate = template.Must(mappingTemplate.Funcs(template.FuncMap{
		"split": split,
	}).Parse(mappingTemplateRaw))

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
	if opts.Host == "" {
		return "", errors.New("host option is required for ambassador 2.0")
	}
	return g.abstractGenerator.Generate(opts, spec)
}

// template func for splitting comma separated strings into an array for iteration
func split(s string, d string) []string {
	if s == "" {
		return []string{}
	}

	return strings.Split(s, d)
}
