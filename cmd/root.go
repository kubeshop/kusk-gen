package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/spec"
)

var (
	apiSpecPath     string
	apiSpecContents *openapi3.T

	serviceName      string
	serviceNamespace string
	servicePort      int32

	rootCmd = &cobra.Command{
		Use:   "kusk",
		Short: "Framework that makes an OpenAPI definition the source of truth for all API-related objects in a cluster (services, mappings, monitors, etc)",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if apiSpecPath == "" {
				log.Fatal(errors.New("no openapi or swagger definition provided"))
			}

			var err error
			apiSpecContents, err = spec.ParseFromFile(apiSpecPath)
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

	rootCmd.PersistentFlags().StringVarP(
		&serviceName,
		"service-name",
		"",
		"",
		"",
	)
	rootCmd.MarkFlagRequired("service-name")

	rootCmd.PersistentFlags().StringVarP(
		&serviceNamespace,
		"service-namespace",
		"",
		"default",
		"",
	)

	rootCmd.PersistentFlags().Int32VarP(
		&servicePort,
		"port",
		"",
		80,
		"",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
