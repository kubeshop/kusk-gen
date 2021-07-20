package generators

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

type Options struct {
	// Namespace for the generated resource. Default value is "default".
	Namespace string

	// Cluster is a set of cluster-wide options
	Cluster *ClusterOptions

	// Service is a set of options of a target service to receive traffic
	Service *ServiceOptions

	// BasePath is the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route).
	// Default value is "/".
	BasePath string

	// TrimPrefix is the prefix that would be omitted from the URL when request is being forwarded
	// to the upstream service, i.e. given that BasePath is set to "/petstore/api/v3", TrimPrefix is set to "/petstore",
	// path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive
	// is "/api/v3/pets".
	TrimPrefix string

	SplitPaths bool
}
