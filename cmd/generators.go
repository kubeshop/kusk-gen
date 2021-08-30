package cmd

import (
	"fmt"
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators"
	_ "github.com/kubeshop/kusk/generators/ambassador"
	_ "github.com/kubeshop/kusk/generators/linkerd"
	_ "github.com/kubeshop/kusk/generators/nginx_ingress"
	"github.com/kubeshop/kusk/options"
	"github.com/kubeshop/kusk/spec"
)

var (
	// this object initially gets filled with whatever we were able to find in x-kusk extension
	// each flag can then override settings
	k = koanf.New(".")

	apiSpecPath string
)

func getOptions() (*options.Options, error) {
	var res options.Options

	err := k.UnmarshalWithConf("", &res, koanf.UnmarshalConf{Tag: "yaml"})
	if err != nil {
		return nil, fmt.Errorf("failed to decode options: %w", err)
	}

	return &res, nil
}

func init() {
	addGenerator := func(gen generators.Interface) {
		cmd := &cobra.Command{
			Use:   gen.Cmd(),
			Short: gen.ShortDescription(),
			Long:  gen.LongDescription(),
			Run: func(cmd *cobra.Command, args []string) {
				if apiSpecPath == "" {
					log.Fatal(fmt.Errorf("no openapi or swagger definition provided"))
				}

				// parse OpenAPI spec
				apiSpec, err := spec.Parse(apiSpecPath)
				if err != nil {
					log.Fatal(err)
				}

				// parse x-kusk top-level extension
				kuskExtensionOpts, err := spec.GetOptions(apiSpec)
				if err != nil {
					log.Fatal(err)
				}

				// populate koanf object with the extension content
				err = k.Load(structs.Provider(*kuskExtensionOpts, "yaml"), nil)
				if err != nil {
					log.Fatal(err)
				}

				// override koanf options with user-provided flags
				err = k.Load(posflag.Provider(cmd.Flags(), ".", k), nil)
				if err != nil {
					log.Fatal(err)
				}

				// fetch merged options
				opts, err := getOptions()
				if err != nil {
					log.Fatal(err)
				}

				opts.PathSubOptions = kuskExtensionOpts.PathSubOptions
				opts.OperationSubOptions = kuskExtensionOpts.OperationSubOptions

				res, err := gen.Generate(opts, apiSpec)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(res)
			},
		}

		addGlobalFlags(cmd)
		// add generator-specific flags
		cmd.Flags().AddFlagSet(gen.Flags())
		cmd.Flags().SortFlags = false

		// register command
		rootCmd.AddCommand(cmd)
	}

	for _, gen := range generators.Registry {
		addGenerator(gen)
	}
}

func addGlobalFlags(cmd *cobra.Command) {
	// add global required flags
	cmd.Flags().StringVarP(
		&apiSpecPath,
		"in",
		"i",
		"",
		"file path to api spec file to generate mappings from. e.g. --in apispec.yaml",
	)
	cmd.MarkFlagRequired("in")

	cmd.Flags().String(
		"namespace",
		"default",
		"namespace for generated resources",
	)

	cmd.Flags().String(
		"service.name",
		"",
		"target Service name",
	)

	cmd.Flags().String(
		"service.namespace",
		"default",
		"namespace containing the target Service",
	)

	cmd.Flags().Int32(
		"service.port",
		80,
		"target Service port",
	)
}
