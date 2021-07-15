package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators/nginx_ingress"
)

var (
	ingressPath          string
	ingressPort          int32
	ingressHost          string
	ingressRewriteTarget string

	// nginxIngressCmd represents the nginxIngress command
	nginxIngressCmd = &cobra.Command{
		Use:   "nginx-ingress",
		Short: "Generates nginx-ingress resources from provided OpenAPI spec",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			ingress, err := nginx_ingress.Generate(&nginx_ingress.Options{
				ServiceName:      serviceName,
				ServiceNamespace: serviceNamespace,
				Host:             ingressHost,
				Path:             ingressPath,
				RewriteTarget:    ingressRewriteTarget,
				Port:             ingressPort,
				TrimPrefix:       trimPrefix,
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
		"/",
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

	nginxIngressCmd.Flags().StringVarP(
		&ingressHost,
		"host",
		"",
		"",
		"",
	)

	nginxIngressCmd.Flags().StringVarP(
		&trimPrefix,
		"trim-prefix",
		"",
		"",
		"",
	)

	nginxIngressCmd.Flags().StringVarP(
		&ingressRewriteTarget,
		"rewrite-target",
		"",
		"",
		"",
	)
}
