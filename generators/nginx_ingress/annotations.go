package nginx_ingress

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kubeshop/kusk/options"
)

const (
	rewriteTargetAnnotationKey = "nginx.ingress.kubernetes.io/rewrite-target"

	// CORS
	corsEnableAnnotationKey = "nginx.ingress.kubernetes.io/enable-cors"

	useRegexAnnotationKey = "nginx.ingress.kubernetes.io/use-regex"
)

func (g *Generator) generateAnnotations(
	path *options.PathOptions,
	nginx *options.NGINXIngressOptions,
	cors *options.CORSOptions,
	timeoutOpts *options.TimeoutOptions,
) map[string]string {
	annotations := map[string]string{}

	if nginx.RewriteTarget != "" {
		annotations[rewriteTargetAnnotationKey] = nginx.RewriteTarget
	} else if len(path.TrimPrefix) > 0 && strings.HasPrefix(path.Base, path.TrimPrefix) {
		annotations[rewriteTargetAnnotationKey] = "/$2"
	}

	// CORS
	if origins := cors.Origins; len(origins) > 0 {
		if len(origins) > 1 {
			log.
				New(os.Stderr, "[WARN]: ", log.Lmsgprefix).
				Printf("Nginx Ingress only supports a single origin. Choosing the first url: %s", origins[0])
		}
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-origin"] = origins[0]
	}

	if methods := cors.Methods; len(methods) > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-methods"] = fmt.Sprintf("%s", strings.Join(methods, ", "))
	}

	if allowHeaders := cors.Headers; len(allowHeaders) > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-headers"] = fmt.Sprintf("%s", strings.Join(allowHeaders, ", "))
	}

	if exposeHeaders := cors.ExposeHeaders; len(exposeHeaders) > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-expose-headers"] = fmt.Sprintf("%s", strings.Join(exposeHeaders, ", "))
	}

	// Default is true, so check if false
	if allowCredentials := cors.Credentials; allowCredentials != nil && !*allowCredentials {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"] = "false"
	}

	if maxAge := cors.MaxAge; maxAge > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-max-age"] = strconv.Itoa(maxAge)
	}
	// End CORS

	// Timeouts
	if requestTimeout := timeoutOpts.RequestTimeout; requestTimeout > 0 {
		strTimeout := strconv.Itoa(int(requestTimeout) / 2)
		if strTimeout == "0" {
			strTimeout = "1"
		}
		annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = strTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = strTimeout
	}
	// End Timeouts

	return annotations
}
