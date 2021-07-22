package generators

import (
	"github.com/go-ozzo/ozzo-validation/v4"
)

type ClusterOptions struct {
	// ClusterDomain is the base DNS domain for the cluster. Default value is "cluster.local".
	ClusterDomain string
}

type ServiceOptions struct {
	// Namespace is the namespace containing the upstream Service.
	Namespace string

	// Name is the upstream Service's name.
	Name string

	// Port is the upstream Service's port. Default value is 80.
	Port int32
}

type PathOptions struct {
	// Base is the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route).
	// Default value is "/".
	Base string

	// TrimPrefix is the prefix that would be omitted from the URL when request is being forwarded
	// to the upstream service, i.e. given that Base is set to "/petstore/api/v3", TrimPrefix is set to "/petstore",
	// path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive
	// is "/api/v3/pets".
	TrimPrefix string

	// Split forces Kusk to generate a separate resource for each Path or Operation, where appropriate.
	Split bool
}

type IngressOptions struct {
	// Host is an ingress host rule.
	// See https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules for additional documentation.
	Host string
}

type NGINXIngressOptions struct {
	// RewriteTarget is a custom rewrite target for ingress-nginx.
	// See https://kubernetes.github.io/ingress-nginx/examples/rewrite/ for additional documentation.
	RewriteTarget string
}

type Options struct {
	// Namespace for the generated resource. Default value is "default".
	Namespace string

	// Service is a set of options of a target service to receive traffic.
	Service ServiceOptions

	// Path is a set of options to configure service endpoints paths.
	Path PathOptions

	// Cluster is a set of cluster-wide options.
	Cluster ClusterOptions

	// Ingress is a set of Ingress-related options.
	Ingress IngressOptions

	// NGINXIngress is a set of custom nginx-ingress options.
	NGINXIngress NGINXIngressOptions
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
}

func (o *Options) Validate() error {
	err := validation.ValidateStruct(o,
		validation.Field(&o.Namespace, validation.Required),
		validation.Field(&o.Service.Name, validation.Required),
	)

	return err
}
