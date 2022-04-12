package flow

import (
	"fmt"
	"log"
	"strconv"

	"github.com/kubeshop/kusk-gen/generators/linkerd"
	"github.com/kubeshop/kusk-gen/options"
)

type linkerdFlow struct {
	baseFlow
}

func (l linkerdFlow) getBasePath() string {
	var basePathSuggestions []string
	for _, server := range l.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	if len(basePathSuggestions) == 0 && l.opts.Path.Base != "" {
		basePathSuggestions = append(basePathSuggestions, l.opts.Path.Base)
	}

	return l.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)
}

func (l linkerdFlow) getClusterDomain() string {
	defaultClusterDomain := "cluster.local"
	if l.opts.Cluster.ClusterDomain != "" {
		defaultClusterDomain = l.opts.Cluster.ClusterDomain
	}

	return l.prompt.InputNonEmpty("Cluster domain", defaultClusterDomain)
}

func (l linkerdFlow) getTimeoutOpts() options.TimeoutOptions {
	var timeoutOptions options.TimeoutOptions

	// Support only request timeout as linkerd generator doesn't support idle timeout
	if requestTimeout := l.prompt.Input("Request timeout, leave empty to skip", strconv.Itoa(int(l.opts.Timeouts.RequestTimeout))); requestTimeout != "" {
		if rTimeout, err := strconv.Atoi(requestTimeout); err != nil {
			log.Printf("WARN: %s is not a valid request timeout value. Skipping\n", requestTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(rTimeout)
		}
	}

	return timeoutOptions
}

func (l linkerdFlow) getCmdFromOpts(opts *options.Options) string {
	cmd := fmt.Sprintf("kusk linkerd -i %s ", l.apiSpecPath)
	cmd = cmd + fmt.Sprintf("--namespace=%s ", l.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.namespace=%s ", l.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.name=%s ", l.targetService)
	cmd = cmd + fmt.Sprintf("--path.base=%s ", opts.Path.Base)
	cmd = cmd + fmt.Sprintf("--cluster.cluster_domain=%s ", opts.Cluster.ClusterDomain)

	if opts.Timeouts.RequestTimeout > 0 {
		cmd = cmd + fmt.Sprintf("--timeouts.request_timeout=%d", opts.Timeouts.RequestTimeout)
	}

	return cmd
}

func (l linkerdFlow) Start() (Response, error) {
	clusterDomain := l.getClusterDomain()

	basePath := l.getBasePath()
	timeoutOptions := l.getTimeoutOpts()

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
		Timeouts: timeoutOptions,
	}

	cmd := l.getCmdFromOpts(opts)

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
