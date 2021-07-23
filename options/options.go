package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Options struct {
	// Namespace for the generated resource. Default value is "default".
	Namespace string `yaml:"namespace"`

	// Service is a set of options of a target service to receive traffic.
	Service ServiceOptions `yaml:"service"`

	// Path is a set of options to configure service endpoints paths.
	Path PathOptions `yaml:"path"`

	// Cluster is a set of cluster-wide options.
	Cluster ClusterOptions `yaml:"cluster"`

	// Ingress is a set of Ingress-related options.
	Ingress IngressOptions `yaml:"ingress"`

	// NGINXIngress is a set of custom nginx-ingress options.
	NGINXIngress NGINXIngressOptions `yaml:"nginx_ingress"`
}

func (o *Options) fillDefaults() {
	if o.Namespace == "" {
		o.Namespace = "default"
	}

	if o.Path.Base == "" {
		o.Path.Base = "/"
	}

	if o.Cluster.ClusterDomain == "" {
		o.Cluster.ClusterDomain = "cluster.local"
	}

	if o.Service.Port == 0 {
		o.Service.Port = 80
	}
}

func (o *Options) Validate() error {
	return v.ValidateStruct(o,
		v.Field(&o.Service, v.Required),
		v.Field(&o.Namespace, v.Required),
	)
}

func (o *Options) FillDefaultsAndValidate() error {
	o.fillDefaults()

	return v.Validate([]v.Validatable{
		o,
		&o.Service,
		&o.Path,
		&o.Cluster,
		&o.Ingress,
		&o.NGINXIngress,
	})
}
