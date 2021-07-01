package ambassador

type Options struct {
	// AmbassadorNamespace is the target namespace for mappings (default: ambassador)
	AmbassadorNamespace string

	ServiceNamespace string
	ServiceName      string

	// BasePath determines the preceding prefix for the route (i.e. /your-prefix/here/rest/of/the/route)
	BasePath string

	// TrimPrefix determines the prefix that would be omitted from the URL when request is being forwarded
	// to the upstream service, i.e. BasePath == /petstore/api/v3, TrimPrefix == /petstore, path that the Mapping would
	// match is /petstore/api/v3/pets, URL that the upstream service would receive is /api/v3/pets
	TrimPrefix string

	// RootOnly determines whether the mappings will be generated for each route (default)
	// or for the root prefix only (requires both BasePath and RootOnly options to be set)
	RootOnly bool
}
