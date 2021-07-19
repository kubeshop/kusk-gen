package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators"
	"github.com/kubeshop/kusk/generators/linkerd"
)

// linkerdCmd represents the linkerd command
var (
	clusterDomain string

	linkerdCmd = &cobra.Command{
		Use:   "linkerd",
		Short: "Generates service profiles for your service",
		Run: func(cmd *cobra.Command, args []string) {
			profiles, err := linkerd.Generate(generators.Options{
				Namespace:              serviceNamespace,
				TargetServiceName:      serviceName,
				TargetServiceNamespace: serviceNamespace,
				Cluster: &generators.ClusterOptions{
					ClusterDomain: clusterDomain,
				},
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

	linkerdCmd.Flags().StringVarP(
		&clusterDomain,
		"cluster-domain",
		"",
		"cluster.local",
		"--cluster-domain cluster.local",
	)
}
