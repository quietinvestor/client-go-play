package pods

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func List(ctx context.Context, clientset *kubernetes.Clientset, namespace string, opts metav1.ListOptions) error {
	logger, _ := logr.FromContext(ctx)
	logger = logger.WithValues("namespace", namespace, "opts", opts)

	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		logger.Error(err, "Failed to list pods")
		return fmt.Errorf("Failed to list pods: %w", err)
	}

	logger.V(2).Info("Successfully listed pods",
		"podCount", len(podList.Items))

	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}

	return nil
}
