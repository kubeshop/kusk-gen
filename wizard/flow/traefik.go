package flow

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kubeshop/kusk-gen/generators/traefik"
	"github.com/kubeshop/kusk-gen/options"
)

type traefikFlow struct {
	baseFlow
}

func (t traefikFlow) getBasePath() string {
	var basePathSuggestions []string
	for _, server := range t.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	if len(basePathSuggestions) == 0 && t.opts.Path.Base != "" {
		basePathSuggestions = append(basePathSuggestions, t.opts.Path.Base)
	}

	return t.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)
}

func (t traefikFlow) getTrimPrefix(basePath string) string {
	trimPrefixDefault := basePath
	if t.opts.Path.TrimPrefix != "" {
		trimPrefixDefault = t.opts.Path.TrimPrefix
	}

	return t.prompt.InputNonEmpty("Prefix to trim from the URL (rewrite)", trimPrefixDefault)
}

func (t traefikFlow) getTimeoutOpts() options.TimeoutOptions {
	var timeoutOptions options.TimeoutOptions

	if requestTimeout := t.prompt.Input("Request timeout, leave empty to skip", strconv.Itoa(int(t.opts.Timeouts.RequestTimeout))); requestTimeout != "" {
		if rTimeout, err := strconv.Atoi(requestTimeout); err != nil {
			log.Printf("WARN: %s is not a valid request timeout value. Skipping\n", requestTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(rTimeout)
		}
	}
	if idleTimeout := t.prompt.Input("Idle timeout, leave empty to skip", strconv.Itoa(int(t.opts.Timeouts.IdleTimeout))); idleTimeout != "" {
		if iTimeout, err := strconv.Atoi(idleTimeout); err != nil {
			log.Printf("WARN: %s is not a valid idle timeout value. Skipping\n", idleTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(iTimeout)
		}
	}

	return timeoutOptions
}

func (t traefikFlow) getCORSOpts() options.CORSOptions {
	var corsOpts options.CORSOptions

	// return if user doesn't want to set CORS options
	if setCORS := t.prompt.Confirm("Set CORS options?"); !setCORS {
		return corsOpts
	}

	// Origins
	// If apispec has some origins set, prompt to use them, else prompt for input
	if len(t.opts.CORS.Origins) > 0 && t.prompt.Confirm(fmt.Sprintf("add the following CORS origins? %s", t.opts.CORS.Origins)) {
		corsOpts.Origins = t.opts.CORS.Origins
	} else {
		corsOpts.Origins = t.prompt.InputMany("add CORS origin")
	}

	// Methods
	// If apispec has some methods set, prompt to use them, else prompt for input
	if len(t.opts.CORS.Methods) > 0 && t.prompt.Confirm(fmt.Sprintf("add the following CORS methods? %s", t.opts.CORS.Methods)) {
		corsOpts.Methods = t.opts.CORS.Methods
	} else {
		corsOpts.Methods = t.prompt.InputMany("add CORS method")
	}

	// Headers
	// If apispec has some headers set, prompt to use them, else prompt for input
	if len(t.opts.CORS.Headers) > 0 && t.prompt.Confirm(fmt.Sprintf("add the following CORS headers? %s", t.opts.CORS.Headers)) {
		corsOpts.Headers = t.opts.CORS.Headers
	} else {
		corsOpts.Headers = t.prompt.InputMany("add CORS header")
	}

	// ExposeHeaders
	// If apispec has some expose headers set, prompt to use them, else prompt for input
	if len(t.opts.CORS.ExposeHeaders) > 0 && t.prompt.Confirm(fmt.Sprintf("add the following CORS expose headers? %s", t.opts.CORS.ExposeHeaders)) {
		corsOpts.ExposeHeaders = t.opts.CORS.ExposeHeaders
	} else {
		corsOpts.ExposeHeaders = t.prompt.InputMany("add CORS headers you want to expose")
	}

	// Credentials
	credentials := t.prompt.Confirm("enable CORS credentials")
	corsOpts.Credentials = &credentials

	// Max age
	// default is what is set in apisec, or 0 if not set
	maxAgeStr := t.prompt.Input("set CORS max age", strconv.Itoa(t.opts.CORS.MaxAge))
	maxAge, err := strconv.Atoi(maxAgeStr)
	if err != nil {
		log.Printf("WARN: %s is not a valid max age value. Skipping\n", maxAgeStr)
		maxAge = 0
	}

	corsOpts.MaxAge = maxAge

	return corsOpts
}

func (t traefikFlow) getCmdFromOpts(opts *options.Options) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("kusk traefik -i %s ", t.apiSpecPath))
	sb.WriteString(fmt.Sprintf("--namespace=%s ", t.targetNamespace))
	sb.WriteString(fmt.Sprintf("--service.namespace=%s ", t.targetNamespace))
	sb.WriteString(fmt.Sprintf("--service.name=%s ", t.targetService))
	sb.WriteString(fmt.Sprintf("--path.base=%s ", opts.Path.Base))

	if opts.Path.TrimPrefix != "" {
		sb.WriteString(fmt.Sprintf("--path.trim_prefix=%s ", opts.Path.TrimPrefix))
	}

	if opts.Host != "" {
		sb.WriteString(fmt.Sprintf("--host=%s ", opts.Host))
	}

	if opts.Timeouts.RequestTimeout > 0 {
		sb.WriteString(fmt.Sprintf("--timeouts.request_timeout=%d", opts.Timeouts.RequestTimeout))
	}
	if opts.Timeouts.IdleTimeout > 0 {
		sb.WriteString(fmt.Sprintf("--timeouts.idle_timeout=%d", opts.Timeouts.IdleTimeout))
	}
	return sb.String()
}

func (t traefikFlow) Start() (Response, error) {
	basePath := t.getBasePath()
	trimPrefix := t.getTrimPrefix(basePath)

	timeoutOptions := t.getTimeoutOpts()
	corsOpts := t.getCORSOpts()

	opts := &options.Options{
		Namespace: t.targetNamespace,
		Service: options.ServiceOptions{
			Namespace: t.targetNamespace,
			Name:      t.targetService,
		},
		Path: options.PathOptions{
			Base:       basePath,
			TrimPrefix: trimPrefix,
		},
		Timeouts: timeoutOptions,
		CORS:     corsOpts,
	}

	cmd := t.getCmdFromOpts(opts)
	var ingressGenerator traefik.Generator
	ingress, err := ingressGenerator.Generate(opts, t.apiSpec)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to generate Ingress: %s\n", err)
	}

	return Response{
		EquivalentCmd: cmd,
		Manifests:     ingress,
	}, nil
}
