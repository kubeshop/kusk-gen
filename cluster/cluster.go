package cluster

import (
	"context"
	"fmt"

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
