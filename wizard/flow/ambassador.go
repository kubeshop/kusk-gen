package flow

import (
	"fmt"

	"github.com/kubeshop/kusk/generators/ambassador"
	"github.com/kubeshop/kusk/options"
)

type ambassadorFlow struct {
	baseFlow
}

func (a ambassadorFlow) Start() (Response, error) {
	var basePathSuggestions []string
	for _, server := range a.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	basePath := a.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)
	trimPrefix := a.prompt.InputNonEmpty("Prefix to trim from the URL (rewrite)", basePath)

	separateMappings := false

	if basePath != "" {
		separateMappings = a.prompt.Confirm("Generate mapping for each endpoint separately?")
	}

	opts := &options.Options{
		Namespace: a.targetNamespace,
		Service: options.ServiceOptions{
			Namespace: a.targetNamespace,
			Name:      a.targetService,
		},
		Path: options.PathOptions{
			Base:       basePath,
			TrimPrefix: trimPrefix,
			Split:      separateMappings,
		},
	}

	cmd := fmt.Sprintf("kusk ambassador -i %s ", a.apiSpecPath)
	cmd = cmd + fmt.Sprintf("--namespace=%s ", a.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.namespace=%s ", a.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.name=%s ", a.targetService)
	cmd = cmd + fmt.Sprintf("--path.base=%s ", basePath)
	if trimPrefix != "" {
		cmd = cmd + fmt.Sprintf("--path.trim_prefix=%s ", trimPrefix)
	}
	if separateMappings {
		cmd = cmd + fmt.Sprintf("--path.split ")
	}

	var ag ambassador.Generator

	mappings, err := ag.Generate(opts, a.apiSpec)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to generate mappings: %s\n", err)
	}

	return Response{
		EquivalentCmd: cmd,
		Manifests:     mappings,
	}, nil
}
