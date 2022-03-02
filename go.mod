module github.com/kubeshop/kusk

go 1.16

require (
	github.com/getkin/kin-openapi v0.64.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/knadh/koanf v1.2.0
	github.com/linkerd/linkerd2 v0.5.1-0.20210701172824-d3cc21da777c
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-isatty v0.0.13
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/traefik/traefik/v2 v2.6.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
)

replace github.com/manifoldco/promptui => github.com/dobegor/promptui v0.8.1-0.20210709101426-a4b2db8f7092

// required by Traefik imports
replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
	github.com/mailgun/multibuf => github.com/containous/multibuf v0.0.0-20190809014333-8b6c9a7e6bba
)
