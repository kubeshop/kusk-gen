package flow

import (
	"fmt"
	"log"
	"strconv"

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

	var timeoutOptions options.TimeoutOptions

	if requestTimeout := a.prompt.Input("Request timeout, leave empty to skip", ""); requestTimeout != "" {
		if rTimeout, err := strconv.Atoi(requestTimeout); err != nil {
			log.Printf("WARN: %s is not a valid request timeout value. Skipping\n", requestTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(rTimeout)
		}
	}

	if idleTimeout := a.prompt.Input("Idle timeout, leave empty to skip", ""); idleTimeout != "" {
		if iTimeout, err := strconv.Atoi(idleTimeout); err != nil {
			log.Printf("WARN: %s is not a valid idle timeout value. Skipping\n", idleTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(iTimeout)
		}
	}

	opts := &options.Options{
		Namespace: a.targetNamespace,
		Service: options.ServiceOptions{
			Namespace: a.targetNamespace,
			Name:      a.targetService,
		},
		Timeouts: timeoutOptions,
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
	if timeoutOptions.RequestTimeout > 0 {
		cmd = cmd + fmt.Sprintf("--timeouts.request_timeout=%d", timeoutOptions.RequestTimeout)
	}
	if timeoutOptions.IdleTimeout > 0 {
		cmd = cmd + fmt.Sprintf("--timeouts.idle_timeout=%d", timeoutOptions.IdleTimeout)
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
