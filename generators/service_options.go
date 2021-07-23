package generators

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type ServiceOptions struct {
	// Namespace is the namespace containing the upstream Service.
	Namespace string `yaml:"namespace"`

	// Name is the upstream Service's name.
	Name string `yaml:"name"`

	// Port is the upstream Service's port. Default value is 80.
	Port int32 `yaml:"port"`
}

func (o *ServiceOptions) Validate() error {
	return v.ValidateStruct(o,
		v.Field(&o.Namespace, v.Required),
		v.Field(&o.Name, v.Required),
		v.Field(&o.Port, v.Required, v.Min(1), v.Max(65535)),
	)
}
