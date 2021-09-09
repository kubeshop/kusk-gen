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
	// Test calling API for Traefik CRD
	t.Run("Traefik CRD API is installed", func(t *testing.T) {
		r := require.New(t)
		traefikClusterDiscovery := fake.NewSimpleClientset()
		traefikClusterDiscovery.Fake.Resources = []*metav1.APIResourceList{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "APIResourceList",
					APIVersion: "v1",
				},
				GroupVersion: "traefik.containo.us/v1alpha1",
				APIResources: []metav1.APIResource{
					{
						Name:               "tlsoptions",
						Namespaced:         true,
						Kind:               "TLSOption",
						Verbs:              metav1.Verbs{"delete", "deletecollection", "get", "list", "patch", "create", "update", "watch"},
						ShortNames:         []string{},
						SingularName:       "tlsoption",
						Categories:         []string{},
						Group:              "",
						Version:            "",
						StorageVersionHash: "",
					},
				},
			},
		}
		client := Client{cs: traefikClusterDiscovery}
		traefikDetected, err := client.DetectTraefikV2()
		r.NoError(err)
		r.Equal(true, traefikDetected)
	})
	// Test calling API when there is no installed CRD.
	t.Run("No Traefik", func(t *testing.T) {
		r := require.New(t)
		client := Client{cs: fake.NewSimpleClientset()}
		traefikDetected, err := client.DetectTraefikV2()
		// Unfortunately fake discovery function doesn't return errors.IsNotFound as real client,
		// so we have to ignore all errors here.
		r.Error(err)
		r.Equal(false, traefikDetected)
	})
}
