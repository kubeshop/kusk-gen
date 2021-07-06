package cmd

import (
	"fmt"
	"log"

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
				Name:          "test",
				Namespace:     "default",
				ClusterDomain: "cluster.local",
			}, apiSpecContents)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(profiles)
		},
	}
)

func init() {
	rootCmd.AddCommand(linkerdCmd)
}
