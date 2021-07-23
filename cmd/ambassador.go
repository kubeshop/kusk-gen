package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators/ambassador"
	"github.com/kubeshop/kusk/options"
)

// ambassadorCmd represents the ambassador command
var (
	ambassadorCmd = &cobra.Command{
		Use:   "ambassador",
		Short: "Generates ambassador mappings for your cluster from the provided api specification",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			generateMappings()
		},
	}

	ambassadorNamespace string
	basePath            string
	trimPrefix          string
	rootOnly            bool
)

func generateMappings() {
	mappings, err := ambassador.Generate(&options.Options{
		Namespace: serviceNamespace,
		Service: options.ServiceOptions{
			Namespace: serviceNamespace,
			Name:      serviceName,
			Port:      servicePort,
		},
		Path: options.PathOptions{
			Base:       basePath,
			TrimPrefix: trimPrefix,
			Split:      !rootOnly,
		},
	}, apiSpec)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mappings)
}

func init() {
	rootCmd.AddCommand(ambassadorCmd)

	// Here you will define your flags and configuration settings.
	// kusk generate -i petstore.yaml --generator=ambassador --service-name=petstore --service-namespace=default --root-only=true

	ambassadorCmd.Flags().StringVarP(
		&ambassadorNamespace,
		"ambassador-namespace",
		"",
		"ambassador",
		"",
	)

	ambassadorCmd.Flags().BoolVarP(
		&rootOnly,
		"root-only",
		"",
		true,
		"",
	)

	ambassadorCmd.Flags().StringVarP(
		&basePath,
		"base-path",
		"",
		"",
		"",
	)

	ambassadorCmd.Flags().StringVarP(
		&trimPrefix,
		"trim-prefix",
		"",
		"",
		"",
	)
}
