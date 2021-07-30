package nginx_ingress

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kubeshop/kusk/options"
)

const (
	rewriteTargetAnnotationKey = "nginx.ingress.kubernetes.io/rewrite-target"

	// CORS
	corsEnableAnnotationKey = "nginx.ingress.kubernetes.io/enable-cors"
)

func (g *Generator) generateAnnotations(options *options.Options) map[string]string {
	annotations := map[string]string{}

	if options.NGINXIngress.RewriteTarget != "" {
		annotations[rewriteTargetAnnotationKey] = options.NGINXIngress.RewriteTarget
	} else if len(options.Path.TrimPrefix) > 0 && strings.HasPrefix(options.Path.Base, options.Path.TrimPrefix) {
		annotations[rewriteTargetAnnotationKey] = "/$2"
	}

	if origins := options.Ingress.CORS.Origins; len(origins) > 0 {
		if len(origins) > 1 {
			log.Printf("[WARN]: Nginx Ingress only supports a single origin. Choosing the first url: %s", origins[0])
		}
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-origin"] = origins[0]
	}

	if methods := options.Ingress.CORS.Methods; len(methods) > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-methods"] = fmt.Sprintf("%s", strings.Join(methods, ", "))
	}

	if allowHeaders := options.Ingress.CORS.Headers; len(allowHeaders) > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-headers"] = fmt.Sprintf("%s", strings.Join(allowHeaders, ", "))
	}

	if exposeHeaders := options.Ingress.CORS.ExposeHeaders; len(exposeHeaders) > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-expose-headers"] = fmt.Sprintf("%s", strings.Join(exposeHeaders, ", "))
	}

	// Default is true, so check if false
	if allowCredentials := options.Ingress.CORS.Credentials; allowCredentials != nil && !*allowCredentials {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"] = "false"
	}

	if maxAge := options.Ingress.CORS.MaxAge; maxAge > 0 {
		annotations[corsEnableAnnotationKey] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-max-age"] = strconv.Itoa(maxAge)
	}

	return annotations
}
