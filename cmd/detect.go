package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"github.com/kubeshop/kusk/cluster"
)

var (
	kubeconfig string

	detectCmd = &cobra.Command{
		Use:   "detect",
		Short: "Connects to current Kubernetes cluster and lists available generators",
		Run: func(cmd *cobra.Command, args []string) {
			detectGenerators()
		},
	}
)

func detectGenerators() {
	client, err := cluster.NewClient(kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	ambassadorDetected, err := client.DetectAmbassador()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Ambassador Edge Stack detected: %v\n", ambassadorDetected)

	linkerdDetected, err := client.DetectLinkerd()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Linkerd detected: %v\n", linkerdDetected)
}

func init() {
	rootCmd.AddCommand(detectCmd)

	detectCmd.Flags().StringVarP(
		&kubeconfig,
		"kubeconfig",
		"",
		filepath.Join(homedir.HomeDir(), ".kube", "config"),
		"path to kubeconfig",
	)
}
