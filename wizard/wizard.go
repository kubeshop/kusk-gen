package wizard

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/client-go/util/homedir"

	"github.com/kubeshop/kusk/cluster"
	"github.com/kubeshop/kusk/spec"
	"github.com/kubeshop/kusk/wizard/flow"
	"github.com/kubeshop/kusk/wizard/prompt"
)

func Start(apiSpecPath string, apiSpec *openapi3.T, prompt prompt.Prompter) {
	canConnectToCluster := false
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if fileExists(kubeConfigPath) {
		fmt.Printf("⎈ kubeconfig detected in %s\n", kubeConfigPath)

		canConnectToCluster = prompt.Confirm(
			"Can Kusk connect to your current cluster to check for supported services and provide suggestions?",
		)
	}

	opts, err := spec.GetOptions(apiSpec)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read options from apispec: %w", err))
	}

	args := &flow.Args{
		ApiSpecPath: apiSpecPath,
		ApiSpec:     apiSpec,
		Prompt:      prompt,
		Opts:        opts,
	}

	var mappings string

	if canConnectToCluster {
		mappings, err = flowWithCluster(args, kubeConfigPath)
	} else {
		mappings, err = flowWithoutCluster(args)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "✔ Done!")

	if prompt.Confirm("Do you want to save mappings to a file (otherwise output to stdout)") {
		saveToPath := prompt.FilePath("Save to", "generated.yaml", false)
		err := os.WriteFile(saveToPath, []byte(mappings), 0666)

		if err != nil {
			log.Fatalf("Failed to save mappings to file: %s\n", err)
		}

		return
	}

	fmt.Println(mappings)
}

func flowWithCluster(args *flow.Args, kubeConfigPath string) (string, error) {
	var servicesToSuggest []string

	client, err := cluster.NewClient(kubeConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to connect to cluster: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Connecting to the cluster...")

	ambassadorFound, err := client.DetectAmbassador()
	if err != nil {
		return "", fmt.Errorf("failed to check if Ambassador is installed: %w", err)
	}

	linkerdFound, err := client.DetectLinkerd()
	if err != nil {
		return "", fmt.Errorf("failed to check if Linkerd is installed: %w", err)
	}

	nginxIngressFound, err := client.DetectNginxIngress()
	if err != nil {
		return "", fmt.Errorf("failed to check if nginx ingress is installed: %w", err)
	}

	traefikFound, err := client.DetectTraefikV2()
	if err != nil {
		return "", fmt.Errorf("failed to check if traefik is installed: %w", err)
	}

	if ambassadorFound {
		servicesToSuggest = append(servicesToSuggest, "ambassador")
		fmt.Fprintln(os.Stderr, "✔ Ambassador installation found")
	}

	if linkerdFound {
		servicesToSuggest = append(servicesToSuggest, "ambassador")
		fmt.Fprintln(os.Stderr, "✔ Linkerd installation found")
	}

	if nginxIngressFound {
		servicesToSuggest = append(servicesToSuggest, "ingress-nginx")
		fmt.Fprintln(os.Stderr, "✔ Ingress Nginx installation found")
	}

	if traefikFound {
		servicesToSuggest = append(servicesToSuggest, "traefik")
		fmt.Fprintln(os.Stderr, "✔ Traefik installation found")
	}

	var targetServiceNamespaceSuggestions []string
	var targetServiceNamespace string

	targetServiceNamespaceSuggestions, err = client.ListNamespaces()
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	targetServiceNamespace = args.Prompt.SelectOneOf("Choose namespace with your service", targetServiceNamespaceSuggestions, true)

	targetServiceSuggestions, err := client.ListServices(targetServiceNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	args.TargetNamespace = targetServiceNamespace

	args.TargetService = args.Prompt.SelectOneOf("Choose your service", targetServiceSuggestions, true)

	args.Service = args.Prompt.SelectOneOf("Choose a service you want Kusk generate manifests for", servicesToSuggest, false)

	return executeFlow(args)
}

func flowWithoutCluster(args *flow.Args) (string, error) {
	defaultNamespace := "default"
	if args.Opts.Service.Namespace != "" {
		defaultNamespace = args.Opts.Service.Namespace
	}
	args.TargetNamespace = args.Prompt.InputNonEmpty("Enter namespace with your service", defaultNamespace)
	args.TargetService = args.Prompt.InputNonEmpty("Enter your service name", args.Opts.Service.Name)

	args.Service = args.Prompt.SelectOneOf(
		"Choose a service you want Kusk generate manifests for",
		[]string{
			"ambassador",
			"linkerd",
			"ingress-nginx",
			"traefik",
		},
		false,
	)

	return executeFlow(args)
}

func executeFlow(args *flow.Args) (string, error) {
	f, err := flow.New(args)
	if err != nil {
		return "", fmt.Errorf("failed to create new flow: %s\n", err)
	}

	response, err := f.Start()
	if err != nil {
		return "", fmt.Errorf("failed to execute flow: %s\n", err)
	}

	if response.EquivalentCmd != "" {
		fmt.Fprintln(os.Stderr, "Here is a CLI command you could use in your scripts (you can pipe it to kubectl):")
		fmt.Fprintln(os.Stderr, response.EquivalentCmd)
	}

	return response.Manifests, nil
}

func fileExists(path string) bool {
	// check if file exists
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}
