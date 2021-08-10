package wizard

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/manifoldco/promptui"
	"k8s.io/client-go/util/homedir"

	"github.com/kubeshop/kusk/cluster"
	"github.com/kubeshop/kusk/generators/ambassador"
	"github.com/kubeshop/kusk/generators/linkerd"
	"github.com/kubeshop/kusk/options"
)

func Start(apiSpecPath string, apiSpec *openapi3.T) {
	canConnectToCluster := false
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if fileExists(kubeConfigPath) {
		fmt.Printf("⎈ kubeconfig detected in %s\n", kubeConfigPath)

		canConnectToCluster = confirm(
			"Can Kusk connect to your current cluster to check for supported services and provide suggestions?",
		)
	}

	var mappings string
	var err error

	if canConnectToCluster {
		mappings, err = flowWithCluster(apiSpecPath, apiSpec, kubeConfigPath)
	} else {
		mappings, err = flowWithoutCluster(apiSpecPath, apiSpec)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "✔ Done!")

	if confirm("Do you want to save mappings to a file (otherwise output to stdout)") {
		saveToPath := promptFilePath("Save to", "generated.yaml", false)
		err := os.WriteFile(saveToPath, []byte(mappings), 0666)

		if err != nil {
			log.Fatalf("Failed to save mappings to file: %s\n", err)
		}

		return
	}

	fmt.Println(mappings)
}

func flowWithCluster(apiSpecPath string, apiSpec *openapi3.T, kubeConfigPath string) (string, error) {
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

	if ambassadorFound {
		servicesToSuggest = append(servicesToSuggest, "ambassador")
		fmt.Fprintln(os.Stderr, "✔ Ambassador installation found")
	}

	if linkerdFound {
		servicesToSuggest = append(servicesToSuggest, "ambassador")
		fmt.Fprintln(os.Stderr, "✔ Linkerd installation found")
	}

	var targetServiceNamespaceSuggestions []string
	var targetServiceNamespace string

	targetServiceNamespaceSuggestions, err = client.ListNamespaces()
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	targetServiceNamespace = selectOneOf("Choose namespace with your service", targetServiceNamespaceSuggestions, true)

	var targetServiceSuggestions []string
	var targetService string

	targetServiceSuggestions, err = client.ListServices(targetServiceNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	targetService = selectOneOf("Choose your service", targetServiceSuggestions, true)

	service := selectOneOf("Choose a service you want Kusk generate manifests for", servicesToSuggest, false)

	switch service {
	case "ambassador":
		return flowAmbassador(apiSpecPath, apiSpec, targetServiceNamespace, targetService)
	case "linkerd":
		return flowLinkerd(apiSpecPath, apiSpec, targetServiceNamespace, targetService)
	}

	return "", fmt.Errorf("unknown service")
}

func flowWithoutCluster(apiSpecPath string, apiSpec *openapi3.T) (string, error) {
	targetServiceNamespace := promptString("Enter namespace with your service", "default")
	targetService := promptString("Enter your service name", "")

	service := selectOneOf("Choose a service you want Kusk generate manifests for", []string{"ambassador", "linkerd"}, false)

	switch service {
	case "ambassador":
		return flowAmbassador(apiSpecPath, apiSpec, targetServiceNamespace, targetService)
	case "linkerd":
		return flowLinkerd(apiSpecPath, apiSpec, targetServiceNamespace, targetService)
	}

	return "", fmt.Errorf("unknown service")
}

func flowAmbassador(apiSpecPath string, apiSpec *openapi3.T, targetNamespace, targetService string) (string, error) {
	var basePathSuggestions []string
	for _, server := range apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	basePath := selectOneOf("Base path prefix", basePathSuggestions, true)
	trimPrefix := promptString("Prefix to trim from the URL (rewrite)", basePath)

	separateMappings := false

	if basePath != "" {
		separateMappings = confirm("Generate mapping for each endpoint separately?")
	}

	opts := &options.Options{
		Namespace: targetNamespace,
		Service: options.ServiceOptions{
			Namespace: targetNamespace,
			Name:      targetService,
		},
		Path: options.PathOptions{
			Base:       basePath,
			TrimPrefix: trimPrefix,
			Split:      separateMappings,
		},
	}

	cmd := fmt.Sprintf("kusk ambassador -i %s ", apiSpecPath)
	cmd = cmd + fmt.Sprintf("--namespace=%s ", targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.namespace=%s ", targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.name=%s ", targetService)
	cmd = cmd + fmt.Sprintf("--path.base=%s ", basePath)
	if trimPrefix != "" {
		cmd = cmd + fmt.Sprintf("--path.trim_prefix=%s ", trimPrefix)
	}
	if separateMappings {
		cmd = cmd + fmt.Sprintf("--path.split ")
	}

	fmt.Fprintln(os.Stderr, "Here is a CLI command you could use in your scripts (you can pipe it to kubectl):")
	fmt.Fprintln(os.Stderr, cmd)

	var ag ambassador.Generator

	mappings, err := ag.Generate(opts, apiSpec)
	if err != nil {
		log.Fatalf("Failed to generate mappings: %s\n", err)
	}

	return mappings, nil
}

func flowLinkerd(apiSpecPath string, apiSpec *openapi3.T, targetNamespace, targetService string) (string, error) {
	clusterDomain := promptString("Cluster domain", "cluster.local")

	opts := &options.Options{
		Namespace: targetNamespace,
		Service: options.ServiceOptions{
			Namespace: targetNamespace,
			Name:      targetService,
		},
		Cluster: options.ClusterOptions{
			ClusterDomain: clusterDomain,
		},
	}

	cmd := fmt.Sprintf("kusk linkerd -i %s ", apiSpecPath)
	cmd = cmd + fmt.Sprintf("--namespace=%s ", targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.namespace=%s ", targetNamespace)
	cmd = cmd + fmt.Sprintf("--service.name=%s ", targetService)
	cmd = cmd + fmt.Sprintf("--cluster.cluster_domain=%s ", clusterDomain)

	fmt.Fprintln(os.Stderr, "Here is a CLI command you could use in your scripts (you can pipe it to kubectl):")
	fmt.Fprintln(os.Stderr, cmd)

	var ld linkerd.Generator

	return ld.Generate(opts, apiSpec)
}

func selectOneOf(label string, variants []string, withAdd bool) string {
	if len(variants) == 0 {
		// it's better to show a prompt
		return promptString(label, "")
	}

	if withAdd {
		p := promptui.SelectWithAdd{
			Label:  label,
			Stdout: os.Stderr,
			Items:  variants,
		}

		_, res, _ := p.Run()
		return res
	}

	p := promptui.Select{
		Label:  label,
		Stdout: os.Stderr,
		Items:  variants,
	}

	_, res, _ := p.Run()
	return res
}

func promptString(label, defaultString string) string {
	p := promptui.Prompt{
		Label:  label,
		Stdout: os.Stderr,
		Validate: func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("should not be empty")
			}

			return nil
		},
		Default: defaultString,
	}

	res, _ := p.Run()

	return res
}

func fileExists(path string) bool {
	// check if file exists
	f, err := os.Stat(path)
	if err == nil && !f.IsDir() {
		return true
	}

	return false
}

func promptFilePath(label, defaultPath string, shouldExist bool) string {
	p := promptui.Prompt{
		Label:   label,
		Stdout:  os.Stderr,
		Default: defaultPath,
		Validate: func(fp string) error {
			if strings.TrimSpace(fp) == "" {
				return errors.New("should not be empty")
			}

			if !shouldExist {
				return nil
			}

			if fileExists(fp) {
				return nil
			}

			return errors.New("should be an existing file")
		},
	}

	res, _ := p.Run()

	return res
}

func confirm(question string) bool {
	p := promptui.Prompt{
		Label:     question,
		Stdout:    os.Stderr,
		IsConfirm: true,
	}

	_, err := p.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false
		}
	}

	return true
}
