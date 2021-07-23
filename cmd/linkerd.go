package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators/linkerd"
	"github.com/kubeshop/kusk/options"
)

// linkerdCmd represents the linkerd command
var (
	clusterDomain string

	linkerdCmd = &cobra.Command{
		Use:   "linkerd",
		Short: "Generates service profiles for your service",
		Run: func(cmd *cobra.Command, args []string) {
			profiles, err := linkerd.Generate(&options.Options{
				Namespace: serviceNamespace,
				Service: options.ServiceOptions{
					Name:      serviceName,
					Namespace: serviceNamespace,
				},
				Cluster: options.ClusterOptions{
					ClusterDomain: clusterDomain,
				},
			}, apiSpec)

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
