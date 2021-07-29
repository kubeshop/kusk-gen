package options

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type IngressOptions struct {
	// Host is an ingress host rule.
	// See https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules for additional documentation.
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	CORS CORSOptions `yaml:"cors,omitempty" json:"cors,omitempty"`
}

type CORSOptions struct {
	Origins        []string `yaml:"origins,omitempty" json:"origins,omitempty"`
	Methods        []string `yaml:"methods,omitempty" json:"methods,omitempty"`
	Headers        []string `yaml:"headers,omitempty" json:"headers,omitempty"`
	ExposedHeaders []string `yaml:"exposed_headers,omitempty" json:"exposed_headers,omitempty"`
	Credentials    bool     `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	MaxAge         int      `yaml:"max_age,omitempty" json:"max_age,omitempty"`
}

func (o *IngressOptions) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Host, is.Host),
	)
}
