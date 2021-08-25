package flow

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kubeshop/kusk/generators/nginx_ingress"
	"github.com/kubeshop/kusk/options"
)

type nginxIngressFlow struct {
	baseFlow
}

func (n nginxIngressFlow) Start() (Response, error) {
	var basePathSuggestions []string
	for _, server := range n.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	basePath := n.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)
	trimPrefix := n.prompt.Input("Prefix to trim from the URL (rewrite)", "")

	separateMappings := false
	if basePath != "" {
		separateMappings = n.prompt.Confirm("Generate ingress resource for each endpoint separately?")
	}

	var timeoutOptions options.TimeoutOptions

	// Support only request timeout as nginx-ingress generator doesn't support idle timeout
	if requestTimeout := n.prompt.Input("Request timeout, leave empty to skip", ""); requestTimeout != "" {
		if rTimeout, err := strconv.Atoi(requestTimeout); err != nil {
			log.Printf("WARN: %s is not a valid request timeout value. Skipping\n", requestTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(rTimeout)
		}
	}

	var corsOpts options.CORSOptions
	if setCORS := n.prompt.Confirm("Set CORS options?"); setCORS {
		// Origins
		corsOpts.Origins = []string{n.prompt.Input("add CORS origin", "")}

		// Methods
		corsOpts.Methods = n.prompt.InputMany("add CORS method")

		// Headers
		corsOpts.Headers = n.prompt.InputMany("add CORS header")

		// ExposeHeaders
		corsOpts.ExposeHeaders = n.prompt.InputMany("add CORS headers you want to expose")

		// Credentials
		credentials := n.prompt.Confirm("enable CORS credentials")
		corsOpts.Credentials = &credentials

		// Max age
		maxAgeStr := n.prompt.Input("set CORS max age", "0")
		maxAge, err := strconv.Atoi(maxAgeStr)
		if err != nil {
			log.Printf("WARN: %s is not a valid max age value. Skipping\n", maxAgeStr)
			maxAge = 0
		}
		corsOpts.MaxAge = maxAge
	}

	opts := &options.Options{
		Namespace: n.targetNamespace,
		Service: options.ServiceOptions{
			Namespace: n.targetNamespace,
			Name:      n.targetService,
		},
		Path: options.PathOptions{
			Base:       basePath,
			TrimPrefix: trimPrefix,
			Split:      separateMappings,
		},
		Timeouts: timeoutOptions,
		CORS:     corsOpts,
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("kusk ambassador -i %s ", n.apiSpecPath))
	sb.WriteString(fmt.Sprintf("--namespace=%s ", n.targetNamespace))
	sb.WriteString(fmt.Sprintf("--service.namespace=%s ", n.targetNamespace))
	sb.WriteString(fmt.Sprintf("--service.name=%s ", n.targetService))
	sb.WriteString(fmt.Sprintf("--path.base=%s ", basePath))

	if trimPrefix != "" {
		sb.WriteString(fmt.Sprintf("--path.trim_prefix=%s ", trimPrefix))
	}

	if separateMappings {
		sb.WriteString("--path.split ")
	}

	var ingressGenerator nginx_ingress.Generator
	ingresses, err := ingressGenerator.Generate(opts, n.apiSpec)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to generate ingresses: %s\n", err)
	}

	return Response{
		EquivalentCmd: sb.String(),
		Manifests:     ingresses,
	}, nil
}
