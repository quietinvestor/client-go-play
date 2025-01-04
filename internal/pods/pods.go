package pods

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PodsList(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]corev1.Pod, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}
