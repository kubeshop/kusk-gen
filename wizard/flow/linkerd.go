package flow

import (
	"fmt"

	"github.com/kubeshop/kusk/generators/linkerd"
	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/wizard/prompt"
)

type linkerdFlow struct {
	baseFlow
}

func (l linkerdFlow) Start() (Response, error) {
	clusterDomain := prompt.InputNonEmpty("Cluster domain", "cluster.local")

	opts := &options.Options{
		Namespace: l.targetNamespace,
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