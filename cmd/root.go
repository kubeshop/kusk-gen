package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kubeshop/openapi-operator/generators/ambassador"
	"github.com/spf13/cobra"
)

var (
	apiSpecPath string

	rootCmd = &cobra.Command{
		Use:   "openapi-operator",
		Short: "Framework that makes an OpenAPI definition the source of truth for all API-related objects in a cluster (services, mappings, monitors, etc)",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			loader := openapi3.Loader{
				Context: context.Background(),
			}

			spec, err := loader.LoadFromFile(apiSpecPath)
			if err != nil {
				log.Fatal(err)
			}

			mappings, err := ambassador.GenerateMappings(ambassador.Options{
				ServiceNamespace: "default",
				ServiceName:      "petstore",
				BasePath:         "/petstore/api/v3",
				TrimPrefix:       "/petstore",
				RootOnly:         true,
			}, spec)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(mappings)
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&apiSpecPath, "in", "i", "", "file path to api spec file to generate mappings from. e.g. -in apispec.yaml")
	rootCmd.MarkFlagRequired("in")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
