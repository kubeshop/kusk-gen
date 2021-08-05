module github.com/kubeshop/kusk

go 1.16

require (
	github.com/getkin/kin-openapi v0.64.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/knadh/koanf v1.2.0
	github.com/linkerd/linkerd2 v0.5.1-0.20210701172824-d3cc21da777c
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-isatty v0.0.13
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
)

replace github.com/manifoldco/promptui => github.com/dobegor/promptui v0.8.1-0.20210709101426-a4b2db8f7092
