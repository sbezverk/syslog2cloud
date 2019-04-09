package utils

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/knative/pkg/apis/duck"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cr "sigs.k8s.io/controller-runtime/pkg/client"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetSinkURI retrieves the sink URI from the object referenced by the given
// ObjectReference.
func GetSinkURI(ctx context.Context, c cr.Client, sink *corev1.ObjectReference, namespace string) (string, error) {
	if sink == nil {
		return "", fmt.Errorf("sink ref is nil")
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(sink.GroupVersionKind())
	err := c.Get(ctx, cr.ObjectKey{Namespace: namespace, Name: sink.Name}, u)
	if err != nil {
		return "", err
	}

	objIdentifier := fmt.Sprintf("\"%s/%s\" (%s)", u.GetNamespace(), u.GetName(), u.GroupVersionKind())

	t := duckv1alpha1.AddressableType{}
	err = duck.FromUnstructured(u, &t)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize sink %s: %v", objIdentifier, err)
	}

	if t.Status.Address == nil {
		return "", fmt.Errorf("sink %s does not contain address", objIdentifier)
	}

	if t.Status.Address.Hostname == "" {
		return "", fmt.Errorf("sink %s contains an empty hostname", objIdentifier)
	}

	return fmt.Sprintf("http://%s/", t.Status.Address.Hostname), nil
}

// NewDefaultClient creates new cloud events client
func NewDefaultClient(target ...string) (client.Client, error) {
	tOpts := []http.Option{http.WithBinaryEncoding()}
	if len(target) > 0 && target[0] != "" {
		tOpts = append(tOpts, http.WithTarget(target[0]))
	}

	// Make an http transport for the CloudEvents client.
	t, err := http.New(tOpts...)
	if err != nil {
		return nil, err
	}
	// Use the transport to make a new CloudEvents client.
	c, err := client.New(t,
		client.WithUUIDs(),
		client.WithTimeNow(),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetClient builds k8s client
func GetClient(kubeCfg string) (*kubernetes.Clientset, *rest.Config, client.Client, error) {
	var config *rest.Config
	var err error
	var kubeconfig string
	kubeconfigEnv := os.Getenv("KUBECONFIG")

	if kubeconfigEnv != "" {
		kubeconfig = kubeCfg
	}
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, nil, nil, err
	}
	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	c, err := client.New(config, client.Options{})
	if err != nil {
		return nil, nil, nil, err
	}

	return k8s, config, c, nil
}
