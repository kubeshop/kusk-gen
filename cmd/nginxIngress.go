package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators/nginx_ingress"
	"github.com/kubeshop/kusk/options"
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
			ingress, err := nginx_ingress.Generate(&options.Options{
				Namespace: serviceNamespace,
				Service: options.ServiceOptions{
					Name:      serviceName,
					Namespace: serviceNamespace,
					Port:      ingressPort,
				},
				Ingress: options.IngressOptions{
					Host: ingressHost,
				},
				Path: options.PathOptions{
					Base:       ingressPath,
					TrimPrefix: trimPrefix,
				},
				NGINXIngress: options.NGINXIngressOptions{
					RewriteTarget: ingressRewriteTarget,
				},
			}, apiSpec)

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
