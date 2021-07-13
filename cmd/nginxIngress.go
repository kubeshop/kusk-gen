package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators/nginxIngress"
)

var (
	ingressPath string
	ingressPort int32

	// nginxIngressCmd represents the nginxIngress command
	nginxIngressCmd = &cobra.Command{
		Use:   "nginx-ingress",
		Short: "Generates nginx-ingress resources from provided OpenAPI spec",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			ingress, err := nginxIngress.Generate(&nginxIngress.Options{
				ServiceName:      serviceName,
				ServiceNamespace: serviceNamespace,
				Path:             ingressPath,
				Port:             ingressPort,
			}, apiSpecContents)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(ingress)
		},
	}
)

func init() {
	rootCmd.AddCommand(nginxIngressCmd)

	nginxIngressCmd.Flags().StringVarP(
		&ingressPath,
		"path",
		"",
		"",
		"",
	)
	nginxIngressCmd.MarkFlagRequired("path")

	nginxIngressCmd.Flags().Int32VarP(
		&ingressPort,
		"port",
		"p",
		80,
		"",
	)
}
