package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk/generators"
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
			ingress, err := nginx_ingress.Generate(&generators.Options{
				Namespace: serviceNamespace,
				Service: &generators.ServiceOptions{
					Name:      serviceName,
					Namespace: serviceNamespace,
					Port:      ingressPort,
				},
				Ingress: &generators.IngressOptions{
					Host: ingressHost,
				},
				Path: &generators.PathOptions{
					Base:       ingressPath,
					TrimPrefix: trimPrefix,
				},
				NGINXIngress: &generators.NGINXIngressOptions{
					RewriteTarget: ingressRewriteTarget,
				},
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
