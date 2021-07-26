package options

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type IngressOptions struct {
	// Host is an ingress host rule.
	// See https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules for additional documentation.
	Host string `yaml:"host,omitempty" json:"host,omitempty"`
}

func (o *IngressOptions) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Host, is.Host),
	)
}
