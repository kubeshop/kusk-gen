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

func (a ambassadorFlow) getBasePath() string {
	var basePathSuggestions []string
	for _, server := range a.apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	if len(basePathSuggestions) == 0 && a.opts.Path.Base != "" {
		basePathSuggestions = append(basePathSuggestions, a.opts.Path.Base)
	}

	return a.prompt.SelectOneOf("Base path prefix", basePathSuggestions, true)
}

func (a ambassadorFlow) getTrimPrefix(basePath string) string {
	trimPrefixDefault := basePath
	if a.opts.Path.TrimPrefix != "" {
		trimPrefixDefault = a.opts.Path.TrimPrefix
	}

	return a.prompt.InputNonEmpty("Prefix to trim from the URL (rewrite)", trimPrefixDefault)
}

func (a ambassadorFlow) shouldSeparateMappings(basePath string) bool {
	return basePath != "" && a.prompt.Confirm("Generate mapping for each endpoint separately?")
}

func (a ambassadorFlow) getTimeoutOpts() options.TimeoutOptions {
	var timeoutOptions options.TimeoutOptions

	if requestTimeout := a.prompt.Input("Request timeout, leave empty to skip", strconv.Itoa(int(a.opts.Timeouts.RequestTimeout))); requestTimeout != "" {
		if rTimeout, err := strconv.Atoi(requestTimeout); err != nil {
			log.Printf("WARN: %s is not a valid request timeout value. Skipping\n", requestTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(rTimeout)
		}
	}

	if idleTimeout := a.prompt.Input("Idle timeout, leave empty to skip", strconv.Itoa(int(a.opts.Timeouts.IdleTimeout))); idleTimeout != "" {
		if iTimeout, err := strconv.Atoi(idleTimeout); err != nil {
			log.Printf("WARN: %s is not a valid idle timeout value. Skipping\n", idleTimeout)
		} else {
			timeoutOptions.RequestTimeout = uint32(iTimeout)
		}
	}

	return timeoutOptions
}

func (a ambassadorFlow) getCORSOpts() options.CORSOptions {
	var corsOpts options.CORSOptions

	// return if user doesn't want to set CORS options
	if setCORS := a.prompt.Confirm("Set CORS options?"); !setCORS {
		return corsOpts
	}

	// Origins
	// If apispec has some origins set, prompt to use them, else prompt for input
	if len(a.opts.CORS.Origins) > 0 && a.prompt.Confirm(fmt.Sprintf("add the following CORS origins? %s", a.opts.CORS.Origins)) {
		corsOpts.Origins = a.opts.CORS.Origins
	} else {
		corsOpts.Origins = a.prompt.InputMany("add CORS origin")
	}

	// Methods
	// If apispec has some methods set, prompt to use them, else prompt for input
	if len(a.opts.CORS.Methods) > 0 && a.prompt.Confirm(fmt.Sprintf("add the following CORS methods? %s", a.opts.CORS.Methods)) {
		corsOpts.Methods = a.opts.CORS.Methods
	} else {
		corsOpts.Methods = a.prompt.InputMany("add CORS method")
	}

	// Headers
	// If apispec has some headers set, prompt to use them, else prompt for input
	if len(a.opts.CORS.Headers) > 0 && a.prompt.Confirm(fmt.Sprintf("add the following CORS headers? %s", a.opts.CORS.Headers)) {
		corsOpts.Headers = a.opts.CORS.Headers
	} else {
		corsOpts.Headers = a.prompt.InputMany("add CORS header")
	}

	// ExposeHeaders
	// If apispec has some expose headers set, prompt to use them, else prompt for input
	if len(a.opts.CORS.ExposeHeaders) > 0 && a.prompt.Confirm(fmt.Sprintf("add the following CORS expose headers? %s", a.opts.CORS.ExposeHeaders)) {
		corsOpts.ExposeHeaders = a.opts.CORS.ExposeHeaders
	} else {
		corsOpts.ExposeHeaders = a.prompt.InputMany("add CORS headers you want to expose")
	}

	// Credentials
	credentials := a.prompt.Confirm("enable CORS credentials")
	corsOpts.Credentials = &credentials

	// Max age
	// default is what is set in apisec, or 0 if not set
	maxAgeStr := a.prompt.Input("set CORS max age", strconv.Itoa(a.opts.CORS.MaxAge))
	maxAge, err := strconv.Atoi(maxAgeStr)
	if err != nil {
		log.Printf("WARN: %s is not a valid max age value. Skipping\n", maxAgeStr)
		maxAge = 0
	}

	corsOpts.MaxAge = maxAge

	return corsOpts
}

func (a ambassadorFlow) getCmdFromOpts(opts *options.Options) string {
	cmd := fmt.Sprintf("kusk ambassador -i %s ", a.apiSpecPath)
	cmd = cmd + fmt.Sprintf("--namespace=%s ", a.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.namespace=%s ", a.targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.name=%s ", a.targetService)
	cmd = cmd + fmt.Sprintf("--path.base=%s ", opts.Path.Base)

	if opts.Path.TrimPrefix != "" {
		cmd = cmd + fmt.Sprintf("--path.trim_prefix=%s ", opts.Path.TrimPrefix)
	}
	if opts.Path.Split {
		cmd = cmd + fmt.Sprintf("--path.split ")
	}

	if opts.Timeouts.RequestTimeout > 0 {
		cmd = cmd + fmt.Sprintf("--timeouts.request_timeout=%d", opts.Timeouts.RequestTimeout)
	}
	if opts.Timeouts.IdleTimeout > 0 {
		cmd = cmd + fmt.Sprintf("--timeouts.idle_timeout=%d", opts.Timeouts.IdleTimeout)
	}

	return cmd
}

func (a ambassadorFlow) Start() (Response, error) {
	basePath := a.getBasePath()
	trimPrefix := a.getTrimPrefix(basePath)
	separateMappings := a.shouldSeparateMappings(basePath)

	timeoutOptions := a.getTimeoutOpts()
	corsOpts := a.getCORSOpts()

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
		CORS: corsOpts,
	}

	cmd := a.getCmdFromOpts(opts)

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
