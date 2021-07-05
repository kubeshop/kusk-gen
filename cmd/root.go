package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

var (
	apiSpecPath     string
	apiSpecContents *openapi3.T

	rootCmd = &cobra.Command{
		Use:   "openapi-operator",
		Short: "Framework that makes an OpenAPI definition the source of truth for all API-related objects in a cluster (services, mappings, monitors, etc)",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if apiSpecPath == "" {
				log.Fatal(errors.New("no openapi or swagger definition provided"))
			}

			loader := openapi3.Loader{
				Context: context.Background(),
			}

			var err error
			apiSpecContents, err = loader.LoadFromFile(apiSpecPath)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&apiSpecPath,
		"in",
		"i",
		"",
		"file path to api spec file to generate mappings from. e.g. -in apispec.yaml",
	)
	rootCmd.MarkFlagRequired("in")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
