package generators

type NGINXIngressOptions struct {
	// RewriteTarget is a custom rewrite target for ingress-nginx.
	// See https://kubernetes.github.io/ingress-nginx/examples/rewrite/ for additional documentation.
	RewriteTarget string `yaml:"rewrite_target"`
}

func (o *NGINXIngressOptions) Validate() error {
	return nil
}
