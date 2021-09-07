package traefik

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
)

var (
	ingressRouteTemplate *template.Template
)

const traefik = "traefik"

func init() {
	generators.Registry[traefik] = &Generator{}

	ingressRouteTemplate = template.Must(template.New("ingressRoute").Parse(ingressRouteTpl))
}

type Generator struct{}

func (g *Generator) Cmd() string {
	return traefik
}

func (g *Generator) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet(traefik, pflag.ExitOnError)
	return fs
}

func (g *Generator) ShortDescription() string {
	return ""
}

func (g *Generator) LongDescription() string {
	return g.ShortDescription()
}

func (g *Generator) Generate(opts *options.Options, spec *openapi3.T) (string, error) {
	if err := opts.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate opts: %w", err)
	}

	var data []ingressRouteData
	op := ingressRouteData{
		Name:             opts.Service.Name,
		Namespace:        opts.Namespace,
		Match:            generateMatchRule(opts),
		ServiceName:      opts.Service.Name,
		ServiceNamespace: opts.Service.Namespace,
		ServicePort:      opts.Service.Port,
	}

	data = append(data, op)

	var buf bytes.Buffer
	err := ingressRouteTemplate.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute ingress route tempalte: %w", err)
	}

	res := buf.String()

	return res, nil
}

func generateMatchRule(opts *options.Options) string {
	return fmt.Sprintf("\"PathPrefix(\\\"%s\\\")\"", opts.Path.Base)
}
