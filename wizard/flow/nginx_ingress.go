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

func (n nginxIngressFlow) getBasePath() string {
	var basePathSuggestions []string
	for _, server := range n.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	if len(basePathSuggestions) == 0 && n.opts.Path.Base != "" {
		basePathSuggestions = append(basePathSuggestions, n.opts.Path.Base)
	}

	return n.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)
}

func (n nginxIngressFlow) getTrimPrefix(basePath string) string {
	trimPrefixDefault := basePath
	if n.opts.Path.TrimPrefix != "" {
		trimPrefixDefault = n.opts.Path.TrimPrefix
	}

	return n.prompt.InputNonEmpty("Prefix to trim from the URL (rewrite)", trimPrefixDefault)
}

func (n nginxIngressFlow) shouldSeparateMappings(basePath string) bool {
	return basePath != "" && n.prompt.Confirm("Generate ingress resource for each endpoint separately?")
}

func (n nginxIngressFlow) getTimeoutOpts() options.TimeoutOptions {
	var timeoutOptions options.TimeoutOptions

	// Support only request timeout as nginx-ingress generator doesn't support idle timeout
	if requestTimeout := n.prompt.Input("Request timeout, leave empty to skip", strconv.Itoa(int(n.opts.Timeouts.RequestTimeout))); requestTimeout != "" {
		if rTimeout, err := strconv.Atoi(requestTimeout); err != nil {
			log.Printf("WARN: %s is not a valid request timeout value. Skipping\n", requestTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(rTimeout)
		}
	}

	return timeoutOptions
}

func (n nginxIngressFlow) getCORSOpts() options.CORSOptions {
	var corsOpts options.CORSOptions

	// return if user doesn't want to set CORS options
	if setCORS := n.prompt.Confirm("Set CORS options?"); !setCORS {
		return corsOpts
	}

	// Origins
	// If apispec has some origins set, prompt to use them, else prompt for input
	if len(n.opts.CORS.Origins) > 0 && n.prompt.Confirm(fmt.Sprintf("add the following CORS origins? %s", n.opts.CORS.Origins)) {
		corsOpts.Origins = n.opts.CORS.Origins
	} else {
		corsOpts.Origins = n.prompt.InputMany("add CORS origin")
	}

	// Methods
	// If apispec has some methods set, prompt to use them, else prompt for input
	if len(n.opts.CORS.Methods) > 0 && n.prompt.Confirm(fmt.Sprintf("add the following CORS methods? %s", n.opts.CORS.Methods)) {
		corsOpts.Methods = n.opts.CORS.Methods
	} else {
		corsOpts.Methods = n.prompt.InputMany("add CORS method")
	}

	// Headers
	// If apispec has some headers set, prompt to use them, else prompt for input
	if len(n.opts.CORS.Headers) > 0 && n.prompt.Confirm(fmt.Sprintf("add the following CORS headers? %s", n.opts.CORS.Headers)) {
		corsOpts.Headers = n.opts.CORS.Headers
	} else {
		corsOpts.Headers = n.prompt.InputMany("add CORS header")
	}

	// ExposeHeaders
	// If apispec has some expose headers set, prompt to use them, else prompt for input
	if len(n.opts.CORS.ExposeHeaders) > 0 && n.prompt.Confirm(fmt.Sprintf("add the following CORS expose headers? %s", n.opts.CORS.ExposeHeaders)) {
		corsOpts.ExposeHeaders = n.opts.CORS.ExposeHeaders
	} else {
		corsOpts.ExposeHeaders = n.prompt.InputMany("add CORS headers you want to expose")
	}

	// Credentials
	credentials := n.prompt.Confirm("enable CORS credentials")
	corsOpts.Credentials = &credentials

	// Max age
	// default is what is set in apisec, or 0 if not set
	maxAgeStr := n.prompt.Input("set CORS max age", strconv.Itoa(n.opts.CORS.MaxAge))
	maxAge, err := strconv.Atoi(maxAgeStr)
	if err != nil {
		log.Printf("WARN: %s is not a valid max age value. Skipping\n", maxAgeStr)
		maxAge = 0
	}

	corsOpts.MaxAge = maxAge

	return corsOpts
}

func (n nginxIngressFlow) getCmdFromOpts(opts *options.Options) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("kusk nginx-ingress -i %s ", n.apiSpecPath))
	sb.WriteString(fmt.Sprintf("--namespace=%s ", n.targetNamespace))
	sb.WriteString(fmt.Sprintf("--service.namespace=%s ", n.targetNamespace))
	sb.WriteString(fmt.Sprintf("--service.name=%s ", n.targetService))
	sb.WriteString(fmt.Sprintf("--path.base=%s ", opts.Path.Base))

	if opts.Path.TrimPrefix != "" {
		sb.WriteString(fmt.Sprintf("--path.trim_prefix=%s ", opts.Path.TrimPrefix))
	}

	if opts.Path.Split {
		sb.WriteString("--path.split ")
	}

	return sb.String()
}

func (n nginxIngressFlow) Start() (Response, error) {
	basePath := n.getBasePath()
	trimPrefix := n.getTrimPrefix(basePath)
	separateMappings := n.shouldSeparateMappings(basePath)

	timeoutOptions := n.getTimeoutOpts()
	corsOpts := n.getCORSOpts()

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

	cmd := n.getCmdFromOpts(opts)

	var ingressGenerator nginx_ingress.Generator
	ingresses, err := ingressGenerator.Generate(opts, n.apiSpec)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to generate ingresses: %s\n", err)
	}

	return Response{
		EquivalentCmd: cmd,
		Manifests:     ingresses,
	}, nil
}
