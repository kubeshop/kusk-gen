package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/interactive"
)

var (
	interactiveCmd = &cobra.Command{
		Use:   "interactive",
		Short: "Connects to current Kubernetes cluster and lists available generators",
		Run: func(cmd *cobra.Command, args []string) {
			interactive.Interactive(apiSpecContents)
		},
	}
)

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
