package flow

import (
	"fmt"

	"github.com/kubeshop/kusk/generators/linkerd"
	"github.com/kubeshop/kusk/options"
)

type linkerdFlow struct {
	baseFlow
}

func (l linkerdFlow) Start() (Response, error) {
	clusterDomain := l.prompt.InputNonEmpty("Cluster domain", "cluster.local")

	var basePathSuggestions []string
	for _, server := range l.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	basePath := l.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)

	opts := &options.Options{
		Namespace: l.targetNamespace,
		Path: options.PathOptions{
			Base: basePath,
		},
		Service: options.ServiceOptions{
			Namespace: l.targetNamespace,
			Name:      l.targetService,
		},
		Cluster: options.ClusterOptions{
			ClusterDomain: clusterDomain,
		},
	}

	cmd := fmt.Sprintf("kusk linkerd -i %s ", l.apiSpecPath)
	cmd = cmd + fmt.Sprintf("--namespace=%s ", l.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.namespace=%s ", l.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.name=%s ", l.targetService)
	cmd = cmd + fmt.Sprintf("--path.base=%s ", basePath)
	cmd = cmd + fmt.Sprintf("--cluster.cluster_domain=%s ", clusterDomain)

	var ld linkerd.Generator

	serviceProfiles, err := ld.Generate(opts, l.apiSpec)
	if err != nil {
		return Response{}, fmt.Errorf("failed to generate linkerd service profiles: %s\n", err)
	}

	return Response{
		EquivalentCmd: cmd,
		Manifests:     serviceProfiles,
	}, nil
}
