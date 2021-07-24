package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/interactive"
	"github.com/kubeshop/kusk/spec"
)

func init() {
	var apiSpecPath string

	interactiveCmd := &cobra.Command{
		Use:   "interactive",
		Short: "Connects to current Kubernetes cluster and lists available generators",
		Run: func(cmd *cobra.Command, args []string) {
			// parse OpenAPI spec
			apiSpec, err := spec.ParseFromFile(apiSpecPath)
			if err != nil {
				log.Fatal(err)
			}

			interactive.Interactive(apiSpec)
		},
	}

	// add common required flags
	interactiveCmd.Flags().StringVarP(
		&apiSpecPath,
		"in",
		"i",
		"",
		"file path to api spec file to generate mappings from. e.g. --in apispec.yaml",
	)
	interactiveCmd.MarkFlagRequired("in")

	rootCmd.AddCommand(interactiveCmd)
}
