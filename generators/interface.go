package generators

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"

	"github.com/kubeshop/kusk-gen/options"
)

type Interface interface {
	Cmd() string
	Flags() *pflag.FlagSet

	ShortDescription() string
	LongDescription() string

	Generate(options *options.Options, spec *openapi3.T) (string, error)
}
