package options

import (
	"github.com/go-ozzo/ozzo-validation/v4"
)

type ClusterOptions struct {
	// ClusterDomain is the base DNS domain for the cluster. Default value is "cluster.local".
	ClusterDomain string `yaml:"cluster_domain"`
}

func (o *ClusterOptions) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.ClusterDomain, validation.Required),
	)
}
