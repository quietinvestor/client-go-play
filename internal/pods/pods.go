package pods

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Interface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error)
}

type Client struct {
	client    kubernetes.Interface
	namespace string
}

func New(client kubernetes.Interface, namespace string) Interface {
	return &Client{
		client:    client,
		namespace: namespace,
	}
}

func (c *Client) List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error) {
	return c.client.CoreV1().Pods(c.namespace).List(ctx, opts)
}
