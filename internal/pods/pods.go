package pods

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func List(ctx context.Context, w io.Writer, clientset *kubernetes.Clientset, namespace string, opts metav1.ListOptions) error {
	logger, _ := logr.FromContext(ctx)
	logger = logger.WithValues("namespace", namespace, "opts", opts)

	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		logger.Error(err, "Failed to list pods")
		return fmt.Errorf("Failed to list pods: %w", err)
	}

	logger.V(2).Info("Successfully listed pods",
		"podCount", len(podList.Items))

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer writer.Flush()

	if namespace == "" {
		fmt.Fprintln(writer, "NAMESPACE\tNAME\tSTATUS\tIP\tNODE")

		for _, pod := range podList.Items {
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", pod.Namespace, pod.Name, pod.Status.Phase, pod.Status.PodIP, pod.Spec.NodeName)
		}

		return nil
	}

	fmt.Fprintln(writer, "NAME\tSTATUS\tIP\tNODE")

	for _, pod := range podList.Items {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", pod.Name, pod.Status.Phase, pod.Status.PodIP, pod.Spec.NodeName)
	}

	return nil
}
