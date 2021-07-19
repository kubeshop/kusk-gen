package generators

type ClusterOptions struct {
	// ClusterDomain is the base DNS domain for the cluster. Default value is "cluster.local".
	ClusterDomain string
}

type Options struct {
	Cluster *ClusterOptions

	// Namespace for the generated resource. Default value is "default".
	Namespace string

	// TargetServiceNamespace is the namespace containing the upstream Service.
	TargetServiceNamespace string

	// TargetServiceNamespace is the upstream Service.
	TargetServiceName string

	TargetServicePort int32

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
