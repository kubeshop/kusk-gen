package interactive

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
)

func Interactive(apiSpec *openapi3.T) {
	var err error

	canConnectToCluster := confirm("Can Kusk connect to your current cluster to check for supported services")

	var client *cluster.Client

	var servicesToSuggest []string

	if canConnectToCluster {
		kubeconfig := promptFilePath("Path to kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), true)

		var err error
		client, err = cluster.NewClient(kubeconfig)
		if err != nil {
			log.Fatalf("Failed to connect to cluster: %s\n", err)
		}

		fmt.Fprintln(os.Stderr, "Connecting to the cluster...")

		ambassadorFound, err := client.DetectAmbassador()
		if err != nil {
			log.Fatalf("Failed to check if Ambassador is installed: %s\n", err)
		}

		linkerdFound, err := client.DetectLinkerd()
		if err != nil {
			log.Fatalf("Failed to check if Linkerd is installed: %s\n", err)
		}

		if ambassadorFound {
			servicesToSuggest = append(servicesToSuggest, "ambassador")
			fmt.Fprintln(os.Stderr, "✔ Ambassador installation found")
		}

		if linkerdFound {
			servicesToSuggest = append(servicesToSuggest, "ambassador")
			fmt.Fprintln(os.Stderr, "✔ Linkerd installation found")
		}
	} else {
		// suggest all services we currently support
		servicesToSuggest = []string{"ambassador", "linkerd"}
	}

	var targetServiceNamespaceSuggestions []string
	var targetServiceNamespace string

	if canConnectToCluster {
		targetServiceNamespaceSuggestions, err = client.ListNamespaces()
		if err != nil {
			log.Fatalf("Failed to list namespaces: %s\n", err)
		}

		targetServiceNamespace = selectOneOf("Choose namespace with your service", targetServiceNamespaceSuggestions, true)
	} else {
		targetServiceNamespace = promptString("Enter namespace with your service", "default")
	}

	var targetServiceSuggestions []string
	var targetService string

	if canConnectToCluster {
		targetServiceSuggestions, err = client.ListServices(targetServiceNamespace)
		if err != nil {
			log.Fatalf("Failed to list namespaces: %s\n", err)
		}

		targetService = selectOneOf("Choose your service", targetServiceSuggestions, true)
	} else {
		targetService = promptString("Enter your service name", "")
	}

	service := selectOneOf("Choose a service you want Kusk generate manifests for", servicesToSuggest, false)

	var mappings string

	switch service {
	case "ambassador":
		mappings, err = flowAmbassador(apiSpec, targetServiceNamespace, targetService)
	case "linkerd":
		mappings, err = flowLinkerd(apiSpec, targetServiceNamespace, targetService)
	default:
		log.Fatal("Unknown service")
	}

	if err != nil {
		log.Fatalf("Failed to generate: %s\n", err)
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

func flowAmbassador(apiSpec *openapi3.T, targetNamespace, targetService string) (string, error) {
	var basePathSuggestions []string
	for _, server := range apiSpec.Servers {
		basePathSuggestions = append(basePathSuggestions, server.URL)
	}

	basePath := selectOneOf("Base path prefix", basePathSuggestions, true)
	trimPrefix := promptString("Prefix to trim from the URL (rewrite)", basePath)

	rootOnly := false

	if basePath != "" {
		rootOnly = confirm("Generate one mapping for all endpoints")
	}

	fmt.Fprintln(os.Stderr, "Generating mappings...")

	mappings, err := ambassador.GenerateMappings(
		ambassador.Options{
			AmbassadorNamespace: "ambassador",
			ServiceNamespace:    targetNamespace,
			ServiceName:         targetService,
			BasePath:            basePath,
			TrimPrefix:          trimPrefix,
			RootOnly:            rootOnly,
		},
		apiSpec,
	)

	if err != nil {
		log.Fatalf("Failed to generate mappings: %s\n", err)
	}

	return mappings, nil
}

func flowLinkerd(apiSpec *openapi3.T, targetNamespace, targetService string) (string, error) {
	clusterDomain := promptString("Cluster domain", "cluster.local")

	return linkerd.Generate(&linkerd.Options{
		ServiceNamespace: targetNamespace,
		ServiceName:      targetService,
		ClusterDomain:    clusterDomain,
	}, apiSpec)
}

func selectOneOf(label string, variants []string, withAdd bool) string {
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

			// check if file exists
			f, err := os.Stat(fp)
			if err == nil && !f.IsDir() {
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
