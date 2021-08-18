package wizard

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/client-go/util/homedir"

	"github.com/kubeshop/kusk/cluster"
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

	var mappings string
	var err error

	if canConnectToCluster {
		mappings, err = flowWithCluster(apiSpecPath, apiSpec, kubeConfigPath, prompt)
	} else {
		mappings, err = flowWithoutCluster(apiSpecPath, apiSpec, prompt)
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

func flowWithCluster(apiSpecPath string, apiSpec *openapi3.T, kubeConfigPath string, prompt prompt.Prompter) (string, error) {
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

	if ambassadorFound {
		servicesToSuggest = append(servicesToSuggest, "ambassador")
		fmt.Fprintln(os.Stderr, "✔ Ambassador installation found")
	}

	if linkerdFound {
		servicesToSuggest = append(servicesToSuggest, "ambassador")
		fmt.Fprintln(os.Stderr, "✔ Linkerd installation found")
	}

	if nginxIngressFound {
		servicesToSuggest = append(servicesToSuggest, "nginx-ingress")
		fmt.Fprintln(os.Stderr, "✔ Nginx Ingress installation found")
	}

	var targetServiceNamespaceSuggestions []string
	var targetServiceNamespace string

	targetServiceNamespaceSuggestions, err = client.ListNamespaces()
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	targetServiceNamespace = prompt.SelectOneOf("Choose namespace with your service", targetServiceNamespaceSuggestions, true)

	targetServiceSuggestions, err := client.ListServices(targetServiceNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	args := &flow.Args{
		ApiSpecPath: apiSpecPath,
		ApiSpec:     apiSpec,
		Prompt:      prompt,
	}

	args.TargetService = prompt.SelectOneOf("Choose your service", targetServiceSuggestions, true)

	args.Service = prompt.SelectOneOf("Choose a service you want Kusk generate manifests for", servicesToSuggest, false)

	return executeFlow(args)
}

func flowWithoutCluster(apiSpecPath string, apiSpec *openapi3.T, prompt prompt.Prompter) (string, error) {
	args := &flow.Args{
		ApiSpecPath: apiSpecPath,
		ApiSpec:     apiSpec,
		Prompt:      prompt,
	}
	args.TargetNamespace = prompt.InputNonEmpty("Enter namespace with your service", "default")
	args.TargetService = prompt.InputNonEmpty("Enter your service name", "")

	args.Service = prompt.SelectOneOf(
		"Choose a service you want Kusk generate manifests for",
		[]string{
			"ambassador",
			"linkerd",
			"nginx-ingress",
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
