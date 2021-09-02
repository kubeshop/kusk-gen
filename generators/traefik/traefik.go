package traefik

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/options"
	"github.com/spf13/pflag"
)

const traefik = "traefik"

func init() {
	generators.Registry[traefik] = &Generator{}
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

func (g *Generator) Generate(options *options.Options, spec *openapi3.T) (string, error) {
	if err := options.FillDefaultsAndValidate(); err != nil {
		return "", fmt.Errorf("failed to validate options: %w", err)
	}

	return "", nil
}
