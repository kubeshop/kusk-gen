package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/version"
)

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints kusk version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Git SHA: %s\n", version.Commit)
			fmt.Printf("Built on: %s\n", version.Date)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
