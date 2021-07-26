package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/spec"
	"github.com/kubeshop/kusk/wizard"
)

func init() {
	var apiSpecPath string

	wizardCmd := &cobra.Command{
		Use:   "wizard",
		Short: "Connects to current Kubernetes cluster and lists available generators",
		Run: func(cmd *cobra.Command, args []string) {
			// parse OpenAPI spec
			apiSpec, err := spec.ParseFromFile(apiSpecPath)
			if err != nil {
				log.Fatal(err)
			}

			wizard.Start(apiSpec)
		},
	}

	// add common required flags
	wizardCmd.Flags().StringVarP(
		&apiSpecPath,
		"in",
		"i",
		"",
		"file path to api spec file to generate mappings from. e.g. --in apispec.yaml",
	)
	wizardCmd.MarkFlagRequired("in")

	rootCmd.AddCommand(wizardCmd)
}
