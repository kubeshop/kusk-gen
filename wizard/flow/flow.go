package flow

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/wizard/prompt"
)

type Interface interface {
	Start() (Response, error)
}

type Response struct {
	EquivalentCmd string
	Manifests     string
}

// Flows "inherit" from this
type baseFlow struct {
	apiSpecPath     string
	apiSpec         *openapi3.T
	targetNamespace string
	targetService   string

	opts *options.Options

	prompt prompt.Prompter
}

type Args struct {
	Service string

	ApiSpecPath     string
	ApiSpec         *openapi3.T
	TargetNamespace string
	TargetService   string

	Opts *options.Options

	Prompt prompt.Prompter
}

// New returns a new flow based on the args.Service
// returns an error if the service isn't supported by a flow
func New(args *Args) (Interface, error) {
	baseFlow := baseFlow{
		apiSpecPath:     args.ApiSpecPath,
		apiSpec:         args.ApiSpec,
		targetNamespace: args.TargetNamespace,
		targetService:   args.TargetService,
		opts:            args.Opts,
		prompt:          args.Prompt,
	}

	switch args.Service {
	case "ambassador":
		return ambassadorFlow{baseFlow}, nil
	case "linkerd":
		return linkerdFlow{baseFlow}, nil
	case "nginx-ingress":
		return nginxIngressFlow{baseFlow}, nil
	case "traefik":
		return traefikFlow{baseFlow}, nil
	default:
		return nil, fmt.Errorf("unsupported service: %s\n", args.Service)
	}
}
