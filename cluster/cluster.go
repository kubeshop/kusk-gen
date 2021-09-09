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
	cs kubernetes.Interface
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

func (c *Client) ListServices(namespace string) ([]string, error) {
	servicesList, err := c.cs.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}

	res := make([]string, len(servicesList.Items))
	for i := range servicesList.Items {
		res[i] = servicesList.Items[i].Name
	}

	return res, nil
}

func (c *Client) ListNamespaces() ([]string, error) {
	namespacesList, err := c.cs.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch namespaces: %w", err)
	}

	res := make([]string, len(namespacesList.Items))
	for i := range namespacesList.Items {
		res[i] = namespacesList.Items[i].Name
	}

	return res, nil
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
		return false, nil
	}

	if partOf, ok := linkerdController.ObjectMeta.Labels["app.kubernetes.io/part-of"]; !ok || partOf != "Linkerd" {
		return false, nil
	}

	return true, nil
}

func (c *Client) DetectNginxIngress() (bool, error) {
	_, err :=
		c.cs.AppsV1().
			Deployments("ingress-nginx").
			Get(context.Background(), "ingress-nginx-controller", metav1.GetOptions{})
	if err == nil {
		return true, nil
	}

	if !errors.IsNotFound(err) {
		return false, fmt.Errorf("error fetching nginx-ingress deployments: %w", err)
	}

	return false, nil
}

func (c *Client) DetectTraefikV2() (bool, error) {
	// We query for resources to check if available API group traefik.containo.us/v1alpha1 is installed
	_, err := c.cs.Discovery().ServerResourcesForGroupVersion("traefik.containo.us/v1alpha1")
	if err == nil {
		return true, nil
	}
	if !errors.IsNotFound(err) {
		return false, fmt.Errorf("error querying cluster for installed CRD API Resources: %w", err)
	}

	return false, nil
}
