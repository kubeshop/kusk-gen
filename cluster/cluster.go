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

	if name, ok := ambassadorDeployment.ObjectMeta.Labels["app.kubernetes.io/name"]; !ok || name != "ambassador" {
		return false, nil
	}

	return true, nil
}

func (c *Client) DetectLinkerd() (bool, error) {

	linkerdController, err :=
		c.cs.AppsV1().
			Deployments("linkerd").
			Get(context.Background(), "linkerd-controller", metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("error fetching Linkerd deployments: %w", err)
	}

	if name, ok := linkerdController.ObjectMeta.Labels["app.kubernetes.io/name"]; !ok || name != "controller" {
		log.Printf("WARN: ")
		return false, nil
	}

	if partOf, ok := linkerdController.ObjectMeta.Labels["app.kubernetes.io/part-of"]; !ok || partOf != "Linkerd" {
		return false, nil
	}

	return true, nil

}
