package cluster

import (
	"testing"

	"github.com/stretchr/testify/require"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClient_ListNamespaces(t *testing.T) {
	data := []struct {
		name               string
		clientset          kubernetes.Interface
		expectedNamespaces []string
	}{
		{
			name:               "no namespaces",
			clientset:          fake.NewSimpleClientset(),
			expectedNamespaces: []string{},
		},
		{
			name: "single namespace",
			clientset: fake.NewSimpleClientset(&v1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}),
			expectedNamespaces: []string{"default"},
		},
	}

	for _, test := range data {
		t.Run(test.name, func(t *testing.T) {
			r := require.New(t)
			client := Client{cs: test.clientset}
			ns, err := client.ListNamespaces()
			r.NoError(err)
			r.Equal(test.expectedNamespaces, ns)
		})
	}
}

func TestClient_ListServices(t *testing.T) {
	data := []struct {
		name             string
		clientset        kubernetes.Interface
		expectedServices []string
	}{
		{
			name:             "no services",
			clientset:        fake.NewSimpleClientset(),
			expectedServices: []string{},
		},
		{
			name: "single service",
			clientset: fake.NewSimpleClientset(&v1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-service",
					Namespace: "default",
				},
				Spec:   v1.ServiceSpec{},
				Status: v1.ServiceStatus{},
			}),
			expectedServices: []string{"some-service"},
		},
	}

	for _, test := range data {
		t.Run(test.name, func(t *testing.T) {
			r := require.New(t)
			client := Client{cs: test.clientset}
			services, err := client.ListServices("default")
			r.NoError(err)
			r.Equal(test.expectedServices, services)
		})
	}
}

func TestClient_DetectAmbassador(t *testing.T) {
	data := []struct {
		name           string
		clientset      kubernetes.Interface
		expectedResult bool
	}{
		{
			name:           "No Ambassador",
			clientset:      fake.NewSimpleClientset(),
			expectedResult: false,
		},
		{
			name: "Invalid Ambassador Spec",
			clientset: fake.NewSimpleClientset(
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "ambassador",
					},
				},
				&apps_v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ambassador",
						Namespace: "ambassador",
					},
				},
			),
			expectedResult: false,
		},
		{
			name: "Valid Ambassador Spec",
			clientset: fake.NewSimpleClientset(
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "ambassador",
					},
				},
				&apps_v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ambassador",
						Namespace: "ambassador",
						Labels: map[string]string{
							"app.kubernetes.io/name": "ambassador",
						},
					},
				},
			),
			expectedResult: true,
		},
	}

	for _, test := range data {
		t.Run(test.name, func(t *testing.T) {
			r := require.New(t)
			client := Client{cs: test.clientset}
			ambassadorDetected, err := client.DetectAmbassador()
			r.NoError(err)
			r.Equal(test.expectedResult, ambassadorDetected)
		})
	}
}

func TestClient_DetectLinkerd(t *testing.T) {
	data := []struct {
		name           string
		clientset      kubernetes.Interface
		expectedResult bool
	}{
		{
			name:           "No Linkerd",
			clientset:      fake.NewSimpleClientset(),
			expectedResult: false,
		},
		{
			name: "Invalid Linkerd Spec",
			clientset: fake.NewSimpleClientset(
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "linkerd",
					},
				},
				&apps_v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "linkerd-controller",
						Namespace: "linkerd",
					},
				},
			),
			expectedResult: false,
		},
		{
			name: "Valid Linkerd Spec",
			clientset: fake.NewSimpleClientset(
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "linkerd",
					},
				},
				&apps_v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "linkerd-controller",
						Namespace: "linkerd",
						Labels: map[string]string{
							"app.kubernetes.io/name":    "controller",
							"app.kubernetes.io/part-of": "Linkerd",
						},
					},
				},
			),
			expectedResult: true,
		},
	}

	for _, test := range data {
		t.Run(test.name, func(t *testing.T) {
			r := require.New(t)
			client := Client{cs: test.clientset}
			linkerdDetected, err := client.DetectLinkerd()
			r.NoError(err)
			r.Equal(test.expectedResult, linkerdDetected)
		})
	}
}

func TestClient_DetectTraefik(t *testing.T) {
	data := []struct {
		name           string
		clientset      kubernetes.Interface
		expectedResult bool
	}{
		{
			name:           "No Traefik",
			clientset:      fake.NewSimpleClientset(),
			expectedResult: false,
		},
		{
			name: "Traefik CRD API is installed",
			clientset: fake.NewSimpleClientset(
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "traefik-v2",
					},
				},
				&apps_v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "traefik",
						Namespace: "traefik-v2",
					},
				},
			),
			expectedResult: true,
		},
	}

	for _, test := range data {
		t.Run(test.name, func(t *testing.T) {
			r := require.New(t)
			client := Client{cs: test.clientset}
			traefikDetected, err := client.DetectTraefikV2()
			r.NoError(err)
			r.Equal(test.expectedResult, traefikDetected)
		})
	}
}
