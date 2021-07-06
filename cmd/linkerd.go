package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators/linkerd"
)

// linkerdCmd represents the linkerd command
var (
	linkerdCmd = &cobra.Command{
		Use:   "linkerd",
		Short: "Generates service profiles for your service",
		Run: func(cmd *cobra.Command, args []string) {
			profiles, err := linkerd.Generate(&linkerd.Options{

			}, apiSpecContents)
		},
	}


)

func init() {
	rootCmd.AddCommand(linkerdCmd)
}
