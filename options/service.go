package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type ServiceOptions struct {
	// Namespace is the namespace containing the upstream Service.
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`

	// Name is the upstream Service's name.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Port is the upstream Service's port. Default value is 80.
	Port int32 `yaml:"port,omitempty" json:"port,omitempty"`
}

type TimeoutOptions struct {
	// RequestTimeout is total request timeout
	RequestTimeout uint32 `yaml:"request_timeout,omitempty" json:"request_timeout,omitempty"`
	// IdleTimeout is timeout for idle connection
	IdleTimeout uint32 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
}

func (o *TimeoutOptions) Validate() error {
	return nil
}

func (o *ServiceOptions) Validate() error {
	return v.ValidateStruct(o,
		v.Field(&o.Namespace, v.Required.Error("service.namespace is required")),
		v.Field(&o.Name, v.Required.Error("service.name is required")),
		v.Field(&o.Port, v.Required.Error("service.port is required"), v.Min(1), v.Max(65535)),
	)
}
