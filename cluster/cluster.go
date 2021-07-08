package cluster

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	cs *kubernetes.Clientset
}

func NewClient(kubeconfig string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// create the client
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{cs}, nil
}

func (c *Client) DetectAmbassador() (bool, error) {
	ambassadorDeployment, err := c.cs.AppsV1().Deployments("ambassador").Get(context.Background(), "ambassador", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("error fetching Ambassador deployment: %w", err)
	}

	_, ambassadorPresent := ambassadorDeployment.ObjectMeta.Labels["app.kubernetes.io/name"]

	return ambassadorPresent, nil
}

func (c *Client) DetectLinkerd() (bool, error) {
	linkerdDeployments, err :=
		c.cs.AppsV1().
			Deployments("linkerd").
			List(
				context.Background(),
				metav1.ListOptions{
					LabelSelector: "app.kubernetes.io/part-of=Linkerd",
				},
			)

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("error fetching Linkerd deployments: %w", err)
	}

	if len(linkerdDeployments.Items) == 0 {
		return false, nil
	}

	if expectedNumberOfLinkerdDeployments := 5; len(linkerdDeployments.Items) < expectedNumberOfLinkerdDeployments {
		log.Printf(
			"number of actual linkerd deployments (%d) less than expected (%d)",
			len(linkerdDeployments.Items),
			expectedNumberOfLinkerdDeployments,
		)
		return false, nil
	}

	linkerdServices, err := c.cs.CoreV1().
		Services("linkerd").
		List(
			context.Background(),
			metav1.ListOptions{
				LabelSelector: "linkerd.io/control-plane-ns=linkerd",
			},
		)

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("error fetching Linkerd services: %w", err)
	}

	if len(linkerdServices.Items) == 0 {
		return false, nil
	}

	if expectedNumberOfLinkerdServices := 7; len(linkerdServices.Items) < expectedNumberOfLinkerdServices {
		log.Printf(
			"number of actual linkerd services (%d) less than expected (%d)",
			len(linkerdServices.Items),
			expectedNumberOfLinkerdServices,
		)
		return false, nil
	}

	return true, nil
}
